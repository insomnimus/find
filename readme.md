Find
---------
A small command line tool that finds files and text inside them.

-	Supports glob patterns for querying files.
-	Incredibly fast, thanks to go's awesome concurrency.
-	Supports regex or plain string matching when querying for text content.
-	1 argument = list matching files
-	2 arguments = find second argument in files matching the first argument.
-	Pass -re to indicate that the second argument is a regex pattern.

Note: Windows users should also use forward slashes when querying.

Note 2: Content querying is case insensitive, glob is not.

Example Usage
-------

	find **/pla*.go "type player struct"

	find -re **/*.go 'func\s?\([^\s]+\s\*?[^\)\)\sFindFile'
