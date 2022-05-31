clear:
	rm -rf ./out

build:
	go build -mod vendor -o out/native/yplw src/*.go

build-x64:
	env GOOS=linux GOARCH=amd64 go build -mod vendor -o out/x64/yplw src/*.go

build-arm7:
	env GOOS=linux GOARCH=arm GOARM=7 go build -mod vendor -o out/arm7/yplw src/*.go

build-deb-x64: build-x64
	rm -rf ./out/x64/deb
	mkdir -p ./out/x64/yplw-pkg/DEBIAN
	mkdir -p ./out/x64/yplw-pkg/usr/bin/
	mkdir -p ./out/x64/yplw-pkg/etc/yplw
	mkdir -p ./out/x64/yplw-pkg/lib/systemd/system/
	cp contrib/deb/control out/x64/yplw-pkg/DEBIAN/control
	cp contrib/config/yplw.toml out/x64/yplw-pkg/etc/yplw/yplw.toml
	cp contrib/lib/systemd/system/yplw.service out/x64/yplw-pkg/lib/systemd/system/yplw.service
	echo "Architecture: amd64" >> out/x64/yplw-pkg/DEBIAN/control
	cp out/x64/yplw out/x64/yplw-pkg/usr/bin/yplw
	dpkg-deb --build out/x64/yplw-pkg

installdeb-x64: clear build-deb-x64
	apt install ./out/x64/yplw-pkg.deb
	systemctl daemon-reload
	systemctl enable yplw
	systemctl start yplw
