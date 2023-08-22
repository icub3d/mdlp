#!/usr/bin/env bash 

# If GOOS is windows, we will use mdlp.exe for the file name.
if [ "$GOOS" == "windows" ]; then
	EXECUTABLE="mdlp.exe"
else
	EXECUTABLE="mdlp"
fi

# If GOOS is linux, tar up the file. Otherwise, zip it.
if [ "$GOOS" == "linux" ]; then
	FILE=$EXECUTABLE-$RELEASE.tar.gz
	rm -f $FILE
	tar -czvf $FILE $EXECUTABLE
else
	FILE=$EXECUTABLE-$RELEASE.zip
	rm -f $FILE
	zip $FILE $EXECUTABLE
fi

# Generate the SHA256 hash of the release.
shasum -a 256 $FILE > $FILE.sha256

# The actual file.
echo curl \
    --fail-with-body -sS \
    -X POST \
    --data-binary @"${FILE}" \
    -H 'Content-Type: application/octet-stream' \
    -H "Authorization: Bearer ${GITHUB_TOKEN}" \
    "${UPLOAD_URL}?name=${FILE}"

# The checksum
echo curl \
    --fail-with-body -sS \
    -X POST \
    --data-binary @"${FILE}.sha256" \
    -H 'Content-Type: text/plain' \
    -H "Authorization: Bearer ${GITHUB_TOKEN}" \
    "${UPLOAD_URL}?name=${FILE}.sha256"
