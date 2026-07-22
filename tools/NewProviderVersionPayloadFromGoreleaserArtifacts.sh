#!/bin/sh
DOWNLOAD_BASE=https://github.com/org/repo/releases/download/
ARTIFACTS_FILE=./dist/artifacts.json
DEFAULT_PROTOCOL_VERSIONS=("5.0", "6.0")

for i in "${DEFAULT_PROTOCOL_VERSIONS[@]}"; do
	DEFAULT_PROTOCOL_VERSIONS_AS_STRING+=$i
done

usage="$(basename "$0") [-h] [-u <base url>] [-a <artifacts file>]-- create new provider version payload from goreleaser artifacts file
    -h  show this help text
    -u  set the base for the download path (default $DOWNLOAD_BASE)
    -a  artifacts file location (default $ARTIFACTS_FILE)
    -p  supported protocol version (multiple can be provided. default "$DEFAULT_PROTOCOL_VERSIONS_AS_STRING")
    "

while getopts 'hu:a:p:' option; do
  case "$option" in
    h) echo "$usage"
       exit
       ;;
    a) ARTIFACTS_FILE=$OPTARG
       ;;
    u) DOWNLOAD_BASE=$OPTARG
       ;;
    p) PROTOCOL_VERSIONS+=("$OPTARG") 
       ;;
   \?) printf "illegal option: -%s\n" "$OPTARG" >&2
       echo "$usage" >&2
       exit 1
       ;;
  esac
done
shift $((OPTIND - 1))

if ! command -v jq &> /dev/null; then 
	echo "error: cannot run because 'jq' is not installed or not available accessible from PATH"
	exit -1
fi;

if [ ! -f $ARTIFACTS_FILE ]; then
	echo "artifacts file ($ARTIFACTS_FILE) not found"
	exit -1
fi;

body=$(jq '.[]  | select(.type=="Archive")' $ARTIFACTS_FILE \
                | jq -n '.artifacts |= [inputs]' \
                | jq '.artifacts | [map(.) | .[] | { os: .goos , arch: .goarch, download_url: ($DOWNLOAD_BASE+ (.path | ltrimstr("dist/"))), shasum: .extra.Checksum | ltrimstr("sha256:") }]'  --arg DOWNLOAD_BASE $DOWNLOAD_BASE \
                | jq -n '.platforms |= inputs' \
                | jq '{"shasums": { "url": ($DOWNLOAD_BASE + $SHASUMURL), "signature_url": ($DOWNLOAD_BASE + $SIGURL) }} + .' --arg SHASUMURL `jq -r '.[] | select(.type=="Checksum") | .name' $ARTIFACTS_FILE` --arg SIGURL `jq -r '.[] | select(.type=="Signature") | .name' $ARTIFACTS_FILE` --arg DOWNLOAD_BASE ${DOWNLOAD_BASE} \
		| jq '{"protocols": [] } + .')

#set the default protocol versions if not specified as command line options
if [ ${#PROTOCOL_VERSIONS[@]} -eq 0 ]; then 
	for i in "${DEFAULT_PROTOCOL_VERSIONS[@]}"; do
		PROTOCOL_VERSIONS+=($i)
	done
fi

#append all the protocol versions to the body
for protocolVersion in "${PROTOCOL_VERSIONS[@]}"; do
	body=$(echo $body | jq '.protocols += [$PROTOCOL_VERSION]' --arg PROTOCOL_VERSION $protocolVersion)
done

echo $body | jq
