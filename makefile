BUILD_DIR := ./build

# all: build_android build_ios

build_android: create_build_dir
	gomobile bind -target=android -androidapi=35 -o ${BUILD_DIR}/android/gitcalendarcore.aar .

build_web: create_build_dir
	cp /opt/homebrew/Cellar/go/1.25.5/libexec/lib/wasm/wasm_exec.js ${BUILD_DIR}/web/
	GOOS=js GOARCH=wasm go build -o ${BUILD_DIR}/web/api.wasm .

# build_macos: create_build_dir
# 	gomobile bind -target=macos .

# build_ios: create_build_dir
# 	gomobile bind -target=ios -o ${BUILD_DIR} .

# ---- helpers ----
create_build_dir:
	mkdir -p ${BUILD_DIR}/android
	mkdir -p ${BUILD_DIR}/web

clean:
	rm -rf ${BUILD_DIR}
