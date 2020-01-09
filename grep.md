
# search all files in s3-bucket/logs with filename starting with "test_log"
snapr grep --s3-dir logs --pattern "(POST.*?00645618)|(PUT.*?00645618)" --s3-key "test_log" 

# search all files in s3-bucket/logs, truncate to 1000 chars
snapr grep --s3-dir logs --pattern "(POST.*?00645618)|(PUT.*?00645618)" --truncate 1000
