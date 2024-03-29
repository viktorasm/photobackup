# S3 backup for photo folders

This tool:
* Goes through selected folder subdirs one by one
* Each subdir is streamed to S3 as zip of the same name as folder, e.g. `2024 01 02 birthday` becomes `2024-01-02-birthday.zip`
* completed folders are skipped in subsequent executions;
* Only supports AWS S3 and selects the "DEEP_ARCHIVE" class (the cheapest one with longest retrieval times). About 1$ per TB/month storage costs at the time of writing.

Writing my own tool was faster than finding something that does what I wanted:)


## TODO:

* extract preview JPEGS from RAWs to keep as separate preview;
* add concurrent uploads;
* change logging to human-friendly version;
* make includes/excludes configurable;