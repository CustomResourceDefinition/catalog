clean: _clean
	docker compose down --remove-orphans --rmi local

_clean:
	-find build -not -name ".gitignore" -and -not -name ".gitkeep" -type f -delete
	-find build -type d -empty -delete
