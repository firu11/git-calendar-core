BUILD_DIR := ./build

# all: build_android build_ios

build_android: create_build_dir
	gomobile bind -target=android -androidapi=35 -o ${BUILD_DIR}/gitcalendarcore.aar .

build_macos: create_build_dir
	gomobile bind -target=macos .

# build_ios: create_build_dir
# 	gomobile bind -target=ios -o ${BUILD_DIR} .

# ---- helpers ----
create_build_dir:
	mkdir -p ${BUILD_DIR}

clean:
	rm -rf ${BUILD_DIR}
