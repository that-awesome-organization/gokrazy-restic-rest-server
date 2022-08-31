all: _gokrazy/extrafiles_arm64.tar _gokrazy/extrafiles_amd64.tar

_gokrazy/extrafiles_amd64.tar:
	mkdir -p _gokrazy/extrafiles_amd64/usr/local/bin
	curl -fsSL https://github.com/restic/rest-server/releases/download/v0.11.0/rest-server_0.11.0_linux_amd64.tar.gz | \
		tar xzv --strip-components=1 -C _gokrazy/extrafiles_amd64/usr/local/bin/ --wildcards \
                rest-server_*/rest-server rest-server_*/LICENSE
	cp dist/blkid.amd64 _gokrazy/extrafiles_amd64/usr/local/bin; mv _gokrazy/extrafiles_amd64/usr/local/bin/LICENSE _gokrazy/extrafiles_amd64/usr/local/bin/LICENSE.rest-server
	cd _gokrazy/extrafiles_amd64 && tar cf ../extrafiles_amd64.tar *
	rm -rf _gokrazy/extrafiles_amd64

_gokrazy/extrafiles_arm64.tar:
	mkdir -p _gokrazy/extrafiles_arm64/usr/local/bin
	curl -fsSL https://github.com/restic/rest-server/releases/download/v0.11.0/rest-server_0.11.0_linux_arm64.tar.gz | \
		tar xzv --strip-components=1 -C _gokrazy/extrafiles_arm64/usr/local/bin/ --wildcards \
		rest-server_*/rest-server rest-server_*/LICENSE
	cp dist/blkid.arm64 _gokrazy/extrafiles_arm64/usr/local/bin; mv _gokrazy/extrafiles_arm64/usr/local/bin/LICENSE _gokrazy/extrafiles_arm64/usr/local/bin/LICENSE.rest-server
	cd _gokrazy/extrafiles_arm64 && tar cf ../extrafiles_arm64.tar *
	rm -rf _gokrazy/extrafiles_arm64

clean:
	rm -f _gokrazy/extrafiles_*.tar
