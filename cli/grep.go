package cli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"snapr/util"
	"strings"

	"github.com/pieterclaerhout/go-waitgroup"
	"github.com/sirupsen/logrus"
)

// GrepCmdRunE runs the grep command
// it is exported for testing
func GrepCmdRunE(ropts *RootCmdOptions, opts *GrepCmdOptions) error {
	funcTag := "grep"
	// logrus.Infof(funcTag)
	// var err error

	// ------  DEFAULTS -----------------------------------

	// ensure ending dir slash for all these
	opts.S3Dir = util.EnsureS3DirPath(opts.S3Dir)

	// if there is a file specified, append it to the dir
	if len(opts.S3Key) > 0 {
		opts.S3Dir = util.JoinS3Path(opts.S3Dir, opts.S3Key)
	}

	// if literal, wrap the regex pattern
	if opts.SearchIsLiteral {
		opts.SearchPattern = "\\b" + opts.SearchPattern + "\\b"
	}

	// ------  VALIDATE -----------------------------------

	// validate the in dir
	if len(opts.S3Dir) == 0 {
		return util.WrapError(fmt.Errorf("validation error"), funcTag, "must provide a value for `--s3-dir`")
	}

	// validate the out dir
	if len(opts.SearchPattern) == 0 {
		return util.WrapError(fmt.Errorf("validation error"), funcTag, "must provide a value for `--pattern`")
	}

	// compile regex
	reggy, err := regexp.Compile(opts.SearchPattern)
	if err != nil {
		return util.WrapError(err, funcTag, "failed to compile search pattern")
	}

	logrus.Infof("IN: %s, OUT: %s, PATTERN: %s", opts.S3Dir, opts.OutDir, opts.SearchPattern)

	// ------  LIST OBJECTS -----------------------------------

	// get a new aws session
	_, s3Client, err := util.NewS3Client(ropts.S3Config)
	if err != nil {
		return util.WrapError(err, funcTag, "failed to get new s3 client")
	}

	// list all SOURCE files recursively
	// for the directory to process from ("originals")
	objectsToProcess, _, err := util.ListS3ObjectsByKey(s3Client, ropts.Bucket, opts.S3Dir, false)
	if err != nil {
		return util.WrapError(err, funcTag, "failed to get src s3 object list")
	}
	logrus.Infof("FILES TO SEARCH: %d", len(objectsToProcess))

	// open a new wait group with a maximum number of concurrent workers
	wg := waitgroup.NewWaitGroup(30)

	// accumulate errors while awaiting
	errorTracker := &[]error{}

	// accumulate results
	resultTracker := &[]*GrepResultChunk{}

	// loop through all objects and spawn goroutines to wait for
	for _, object := range objectsToProcess {

		// block adding until the next worker has finished
		wg.BlockAdd()

		// on a separate goroutine, do something asyncronous
		// download, search, accumulate
		go func(searchObj *util.S3Object, accumulator *[]*GrepResultChunk, errorAccumulator *[]error) {
			funcTag := "GrepSearchWorker"
			defer wg.Done()

			logrus.Infof("SEARCH KEY: %s", searchObj.Key)

			// download the original file
			dlBytes, err := util.DownloadS3Object(s3Client, ropts.Bucket, searchObj.Key)
			if err != nil {
				err = util.WrapError(err, funcTag, fmt.Sprintf("failed to download bucket object: %s", searchObj.Key))
				logrus.Warnf(err.Error())
				*errorAccumulator = append(*errorAccumulator, err)
			}

			// encode for escape chars
			buf := new(bytes.Buffer)
			enc := json.NewEncoder(buf)
			enc.SetEscapeHTML(false)
			err = enc.Encode(string(dlBytes))
			if err != nil {
				err = util.WrapError(err, funcTag, fmt.Sprintf("failed to convert downloaded bytes to string: %s", searchObj.Key))
				logrus.Warnf(err.Error())
				*errorAccumulator = append(*errorAccumulator, err)
			}
			logrus.Infof("LENGTH: %d, KEY: %s", buf.Len(), searchObj.Key)

			// open a new scanner
			scanr := bufio.NewScanner(buf)
			// scanr := bufio.NewScanner(strings.NewReader(buf.String()))

			// use a custom split method and buffer
			scanrBufLimit := 1024 * 1024
			scanrBuf := make([]byte, 0, scanrBufLimit)
			scanr.Buffer(scanrBuf, 10*scanrBufLimit)

			// scan lines (not words or other)
			// in our case, we have one HUGE line
			scanr.Split(bufio.ScanLines)

			// scan and search
			lineCounter := 0
			lineMatchCounter := 0
			matchCounter := 0
			for scanr.Scan() {

				// tick the line counter
				lineCounter++

				// scan the line
				line := scanr.Text()
				// logrus.Infof("LINE: %d, LEN: %d", lineCounter, len(line))

				// if matches, look closer
				if reggy.MatchString(line) {

					// reset counters
					lineMatchCounter++

					// check for sub matches in the line
					results := reggy.FindAllStringIndex(line, -1)
					for _, r := range results {
						matchCounter++

						// raw indexes and text
						startIdx := r[0]
						endIdx := r[1]
						rawText := line[startIdx:endIdx]

						// display x chars before and after each match
						displayOffset := 50
						displayStart := 0
						if startIdx > displayOffset {
							displayStart = startIdx - displayOffset
						}
						displayEnd := len(line) - 1
						if endIdx < len(line)-1-displayOffset {
							displayEnd = endIdx + displayOffset
						}
						displayText := line[displayStart:displayEnd]

						// truncation
						truncLength := opts.TruncationLimit
						var truncTexts []string

						// if less than max, just return that
						if len(rawText) <= truncLength {
							truncTexts = append(truncTexts, rawText)
						} else {
							// if length is longer than max, show beginning and ending with ... in middle
							truncTexts = append(truncTexts, line[startIdx:startIdx+(truncLength/2)])
							truncTexts = append(truncTexts, line[(endIdx-(truncLength/2)):endIdx])
						}

						// logrus.Infof("(MATCH %d) Line: %d, Start: %d, End: %d", matchCounter, lineCounter, startIdx, endIdx)
						//logrus.Infof("MATCH: %s", line[displayStart:displayEnd])

						*accumulator = append(*accumulator, &GrepResultChunk{
							S3Object:      searchObj,
							LineNumber:    lineCounter,
							StartIndex:    startIdx,
							EndIndex:      endIdx,
							RawText:       rawText,
							DisplayText:   displayText,
							TruncatedText: strings.Join(truncTexts, " ... "),
						})
					}
				}
			}

			// we need these injected here
		}(object, resultTracker, errorTracker)
	}

	// wait on everything to complete
	wg.Wait()

	// further filters or scoring systems?

	logrus.Infof("FOUND: %d", len(*resultTracker))

	logrus.Infof("======================================")
	for i, r := range *resultTracker {
		logrus.Infof("(RESULT %d)", i+1)
		logrus.Infof("%s:%d[%d]", r.S3Object.Key, r.LineNumber, r.StartIndex)
		logrus.Infof("(LEN %d) %s", len(r.RawText), r.TruncatedText)
		// logrus.Infof("%s", r.DisplayText)
		logrus.Infof("======================================")
	}
	logrus.Infof("TOTAL: %d", len(*resultTracker))

	return nil
}

// GrepResultChunk holds a reference to everything we need to identify a match
type GrepResultChunk struct {
	S3Object      *util.S3Object
	LineNumber    int
	StartIndex    int
	EndIndex      int
	DisplayText   string
	TruncatedText string
	RawText       string
}
