# jstests

This is a directory of js files that will be run as part of the integrationtest suite. 
These files will be run individually, the equivalent of `mongo localhost:27016 <FILENAME>`, so for each file we should make sure to:
- use unique data: as each test will be run against the same DB we want to avoid data conflicts
- do your own feature discovery: if you need to run a test only if a certain feature is enabled; check your conditions before running the test
- do asserts: the files are just run end-to-end; it is up the script to do the appropriate asserts
