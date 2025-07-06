7z a -tzip "bundled.zip" "./bundled/*" -mx5
rsrc -manifest main.manifest -o rsrc.syso
go build -ldflags="-H windowsgui"
rm rsrc.syso
rm bundled.zip