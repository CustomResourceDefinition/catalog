clean:
	-find build -type f ! -name ".gitignore" ! -name ".gitkeep" ! -path "build/remote/*" ! -path "build/remote" -delete
	-find build -type d -empty -delete
