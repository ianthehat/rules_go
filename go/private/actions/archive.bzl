# Copyright 2014 The Bazel Authors. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

load("@io_bazel_rules_go//go/private:common.bzl",
    "split_srcs",
)
load("@io_bazel_rules_go//go/private:mode.bzl",
    "get_mode",
    "mode_string",
)
load("@io_bazel_rules_go//go/private:providers.bzl",
    "GoLibrary",
    "GoEmbed",
    "GoArchive",
)

def get_archive(lib, mode):
  for a in lib.archives:
    if a.mode == mode:
      return a
  if a.mode != mode: fail("No archive for {} matching {}".format(lib.label, mode_string(mode)))

def emit_archive(ctx, go_toolchain, mode=None, importpath=None, goembed=None, direct=None, importable=True):
  """See go/toolchains.rst#archive for full documentation."""

  if importpath == None: fail("importpath is a required parameter")
  if goembed == None: fail("goembed is a required parameter")
  if mode == None: fail("mode is a required parameter")

  source = split_srcs(goembed.build_srcs)
  lib_name = importpath + ".a"
  compilepath = importpath if importable else None
  out_dir = "~{}~{}~".format(mode_string(mode), ctx.label.name)
  out_lib = ctx.actions.declare_file("{}/{}".format(out_dir, lib_name))
  searchpath = out_lib.path[:-len(lib_name)]

  extra_objects = []
  for src in source.asm:
    obj = ctx.actions.declare_file("{}/{}.o".format(out_dir, src.basename[:-2]))
    go_toolchain.actions.asm(ctx, go_toolchain, mode=mode, source=src, hdrs=source.headers, out_obj=obj)
    extra_objects += [obj]
  archive = goembed.cgo_info.archive if goembed.cgo_info else None

  transitive = depset()
  for a in direct:
    transitive += [a]
    transitive += a.transitive
  for a in transitive:
    if a.mode != mode: fail("Archive mode does not match {} is {} expected {}".format(importpath, mode_string(a.mode), mode_string(mode)))

  if len(extra_objects) == 0 and archive == None:
    go_toolchain.actions.compile(ctx,
        go_toolchain = go_toolchain,
        sources = source.go,
        importpath = compilepath,
        archives = direct,
        mode = mode,
        out_lib = out_lib,
        gc_goopts = goembed.gc_goopts,
    )
  else:
    partial_lib = ctx.actions.declare_file("{}/~partial.a".format(out_dir))
    go_toolchain.actions.compile(ctx,
        go_toolchain = go_toolchain,
        sources = source.go,
        importpath = compilepath,
        archives = direct,
        mode = mode,
        out_lib = partial_lib,
        gc_goopts = goembed.gc_goopts,
    )
    go_toolchain.actions.pack(ctx,
        go_toolchain = go_toolchain,
        mode = mode,
        in_lib = partial_lib,
        out_lib = out_lib,
        objects = extra_objects,
        archive = archive,
    )
  return GoArchive(
      mode = mode,
      file = out_lib,
      importpath = importpath,
      searchpath = searchpath,
      embed = goembed,
      direct = direct,
      transitive = transitive,
  )
