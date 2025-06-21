clean: _clean
	docker compose down --remove-orphans --volumes --rmi local

_clean:
	rm -r build/bin build/ephemeral/templates/* build/ephemeral/schema/* &>/dev/null || true
