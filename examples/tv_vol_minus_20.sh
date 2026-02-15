#!/usr/bin/env bash
# Paulo Aleixo Campos
__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
__dbg_on_off=off  # on off
function shw_info { echo -e '\033[1;34m'"$1"'\033[0m'; }
function error { echo "ERROR in ${1}"; exit 99; }
trap 'error $LINENO' ERR
function dbg { [[ "$__dbg_on_off" == "on" ]] || return; echo -e '\033[1;34m'"dbg $(date +%Y%m%d%H%M%S) ${BASH_LINENO[0]}\t: $@"'\033[0m';  }
#exec > >(tee -i /tmp/$(date +%Y%m%d%H%M%S.%N)__$(basename $0).log ) 2>&1
set -o errexit
  # NOTE: the "trap ... ERR" alreay stops execution at any error, even when above line is commente-out
set -o pipefail
set -o nounset
set -o xtrace

SOCKS5_PROXY="192.168.255.5:1080"

cd "$__dir"/..
# Obtener volumen actual
current_vol=$(./lg-webos-ssap -use-socks5-proxy "$SOCKS5_PROXY" -cmd vol-get | jq -r '.volume')
  # {
  #   "muted": false,
  #   "returnValue": true,
  #   "scenario": "mastervolume_tv_speaker",
  #   "volume": 7
  # }


# Calcular nuevo volumen (restar 20)
if [ -z "$current_vol" ]; then
    echo "No se pudo obtener el volumen actual. Asumiendo 20 por defecto."
    new_vol=0
else
    new_vol=$((current_vol - 20))
fi

# Asegurarse de que no sea negativo
if [ "$new_vol" -lt 0 ]; then
    new_vol=0
fi

# Establecer nuevo volumen
./lg-webos-ssap -use-socks5-proxy "$SOCKS5_PROXY" -cmd vol-set -arg "$new_vol"
