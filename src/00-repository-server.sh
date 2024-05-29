cd /repository;
python3 -m http.server 2>&1 | grep -v '" 200 -' | grep -v 'Connection refused' &
