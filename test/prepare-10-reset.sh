printf "Reset for a fresh test ... "
rm -rf /schema/* &>/dev/null || true
rm -rf /templates/* &>/dev/null || true
rm -rf /root/.cache/helm/* &>/dev/null || true
rm -rf /root/.config/helm/* &>/dev/null || true
rm -rf /root/.local/helm/* &>/dev/null || true
rm -rf /repository/* &>/dev/null || true
echo "done"
echo
