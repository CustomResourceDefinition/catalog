clean: _clean
	docker compose down --remove-orphans --volumes --rmi local

_clean:
	-rm -fr build/bin build/ephemeral/templates/* build/ephemeral/schema/*
