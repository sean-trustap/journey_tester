# Copyright 2026 Trustap. All rights reserved.
# Use of this source code is governed by an MIT
# licence that can be found in the LICENCE file.

tgt_dir:=target
tgt_artfs_dir:=$(tgt_dir)/artfs
tgt_deps_dir:=$(tgt_dir)/deps
tgt_gen_dir:=$(tgt_dir)/gen
tgt_tmp_dir:=$(tgt_dir)/tmp
tgt_img_ctx_dir:=$(tgt_tmp_dir)/img_ctx
tgt_test_dir:=$(tgt_tmp_dir)/test
tgt:=$(tgt_artfs_dir)/tgt_name

# Aliases for automatic variables, see `docs/makefiles.md` for more details.
target=$@
first_dep=$<
all_deps=$^
directory_dep=$|
stem=$*

$(tgt_gen_dir)/test_server: \
		tools/test_server/main.go \
		| $(tgt_gen_dir)
	@# We disable cgo when building so that these executables can be run in
	@# environments outside those that they were built in. For example, builds
	@# are generally performed in a Debian environment but executables will
	@# ideally run in minimal Alpine environments. Note that this can result in
	@# slower builds (<https://stackoverflow.com/a/47715985>).
	( \
		cd '$(dir $(first_dep))' \
			&& go get -v \
			&& CGO_ENABLED=0 go build \
				-o '$(CURDIR)/$(target)' \
	)

$(tgt_artfs_dir): | $(tgt_dir)
	mkdir '$(target)'

$(tgt_gen_dir): | $(tgt_dir)
	mkdir '$(target)'

$(tgt_tmp_dir): | $(tgt_dir)
	mkdir '$(target)'

$(tgt_test_dir): | $(tgt_tmp_dir)
	mkdir '$(target)'

$(tgt_dir):
	mkdir '$(target)'
