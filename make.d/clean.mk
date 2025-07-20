clean: _clean
	docker compose down --remove-orphans --rmi local

_clean:
	-find build -type f ! -name ".gitignore" ! -name ".gitkeep" ! -path "build/remote/*" ! -path "build/remote" -delete
	-find build -type d -empty -delete
