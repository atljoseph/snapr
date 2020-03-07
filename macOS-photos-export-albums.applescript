
-- export albums to folders and rename the output with the title if set,
-- instead of using the original filename

-- get the dest path from the user
-- set dest to "/Users/joseph/Desktop/PHOTOS-Albums/" as POSIX file as text
set destAlias to choose folder with prompt "Choose an output location:"
set destPath to destAlias as string


-- get a list of albums and export each
tell application "Photos"
	
	-- bring photos to foreground
	-- activate
	
	-- get a list of all photo albums from the app
	set albList to name of albums
	
	-- COMMENT THIS OUT TO DO ALL ALBUMS THAT START WITH "web-"
	-- prompt user for albums
	set albList to choose from list albList with prompt "Select some albums" with multiple selections allowed
	
	-- if valid (not cancelled, or empty list)
	if albList is not false then
		
		-- automatically go through and do all albums meeting criteria
		repeat with fullAlbName in albList
			log fullAlbName
			
			-- only do certain albums
			if fullAlbName starts with "web-" then
				
				-- strip off the prefix
				set albName to text 5 thru -1 of fullAlbName
				
				-- determine the foldername
				set albFolder to destPath & albName
				
				-- remove the output folder if it exists
				my rmDir(albFolder)
				
				-- with a timeout of 20 minutes - needed for some of the larger albums
				with timeout of 1200 seconds
					try
						-- determine the foldername
						set albFolder to destPath & albName
						
						-- create a folder named (the name of this album) in dest
						my mkdir(albFolder)
						log "done with mkdir: " & albName
						
						-- get media items
						set mediaItems to get media items of album fullAlbName
						
						-- export the files
						-- WILL ALWAYS OUTPUT AS ".JPG"
						log "started export for " & albName
						export mediaItems to (albFolder as alias) without using originals
						log "done exporting album: " & albFolder
						
					on error the errorMessageExport
						error "Error exporting: " & errorMessageExport
						quit
					end try
				end timeout
				
				-- file to track image names
				set imagListFileName to "album.txt"
				tell application "Finder"
					make new file at albFolder with properties {name:imagListFileName, file type:"TEXT", creator type:"ttxt"}
					set trackerFile to file (albFolder & ":" & imagListFileName)
				end tell
				
				-- after export
				-- set all the outputted file's names to the image title
				-- retain the extensions
				try
					repeat with im in mediaItems
						
						try
							-- export command ALWAYS exports as JPG for some reason
							set newFileName to my renameOutput(albFolder, name of im, filename of im, "jpg")
							
							--add line to text file
							-- tell application "TextEdit" to make new paragraph at after last character of text of trackerFile with data "This is the newcomer sentence."
							set written to my write_to_file(trackerFile, albName & "/" & newFileName & "
", 1)
						on error
							-- add a line with the original name of the file instead
							set written to my write_to_file(trackerFile, albName & "/" & (filename of im) & "
", 1)
						end try
						
					end repeat
					log "done renaming outputs"
					
				on error the errorMessageRename
					error "Error renaming: " & errorMessageRename
					quit
				end try
				
				
			end if
		end repeat
	end if
	
	
end tell


-- functions
on write_to_file(target_file, this_data, append_data)
	try
		set the target_file to the target_file as string
		set the open_target_file to open for access file target_file with write permission
		write this_data to the open_target_file starting at eof
		close access the open_target_file
		return true
	on error
		try
			close access file target_file
		end try
		return false
	end try
end write_to_file
on renameOutput(baseFolder, photoTitle, photoFilename, targetExtension)
	log baseFolder
	
	-- use finder
	-- get os ref to each file and rename
	tell application "Finder"
		
		-- original file name and extension
		set lastDotIdx to my last_offset(photoFilename, ".")
		set inFileNameNoExt to text 1 thru (lastDotIdx - 1) of photoFilename
		set inFileNameBeforeExt to text lastDotIdx thru -1 of photoFilename
		set inFileName to inFileNameNoExt & "." & targetExtension
		log "photoFilename: " & photoFilename
		-- log "inFileNameNoExt: " & inFileNameNoExt
		-- log "inFileNameBeforeExt: " & inFileNameBeforeExt
		
		-- original file path
		set inFilePath to baseFolder & ":" & inFileName
		set inFile to file inFilePath
		
		set defaultFileName to inFileNameNoExt & "." & targetExtension
		log "defaultFileName: " & defaultFileName
		
		-- only rename if the title is not empty
		if (photoTitle is not missing value) then
			
			-- needed the extra block here because short circuit did not work
			if (length of photoTitle > 0) then
				
				-- new file name
				set outFileName to photoTitle & "." & targetExtension
				log (inFileName & " --> " & outFileName)
				
				-- perform a rename operation
				-- set name of fpOrig to fpNew
				set name of file inFileName of folder baseFolder to outFileName
				
				log "new photo title: " & outFileName
				return outFileName
			else
				log "photo has empty title: " & photoFilename
				return defaultFileName
			end if
		else
			log "photo does not have title defined: " & photoFilename
			return defaultFileName
		end if
		
		-- dummy return for function to work
		return defaultFileName
		
	end tell
end renameOutput
on rmDir(rmPath)
	tell application "Finder"
		if exists folder rmPath then
			-- delete (files of rmPath)
			log ("deleting " & rmPath)
			delete rmPath
		end if
	end tell
end rmDir
on mkdir(mkdirPath)
	do shell script "mkdir -p " & quoted form of POSIX path of mkdirPath
end mkdir
on last_offset(the_text, char)
	try
		set len to count of the_text
		set reversed to reverse of characters of the_text as string
		set last_occurrence to len - (offset of char in reversed) + 1
		if last_occurrence > len then
			return 0
		end if
	on error
		return 0
	end try
	return last_occurrence
end last_offset