# Copyright 2017 The Bazel Authors. All rights reserved.
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

# Modes are documented in go/modes.rst#compilation-modes

LINKMODE_NORMAL = "normal"
LINKMODE_SHARED = "shared"
LINKMODE_PIE = "pie"
LINKMODE_PLUGIN = "plugin"

def mode_string(mode):
  result = []
  if mode.static:
    result.append("static")
  if mode.race:
    result.append("race")
  if mode.msan:
    result.append("msan")
  if mode.pure:
    result.append("pure")
  if mode.debug:
    result.append("debug")
  if mode.strip:
    result.append("stripped")
  if not result or not mode.link == LINKMODE_NORMAL:
    result.append(mode.link)
  return "_".join(result)

def _ternary(*values):
  for v in values:
    if v == None: continue
    if type(v) == "bool": return v
    if type(v) != "string": fail("Invalid value type {}".format(type(v)))
    v = v.lower()
    if v == "on": return True
    if v == "off": return False
    if v == "auto": continue
    fail("Invalid value {}".format(v))
  fail("_ternary failed to produce a final result from {}".format(values))

def get_mode(ctx, toolchain_flags):
  force_pure = None
  if "@io_bazel_rules_go//go:toolchain" in ctx.toolchains:
    if ctx.toolchains["@io_bazel_rules_go//go:toolchain"].cross_compile:
      # We always have to user the pure stdlib in cross compilation mode
      force_pure = True

  #TODO: allow link mode selection
  debug = False
  strip = True
  if toolchain_flags:
    debug = toolchain_flags.compilation_mode == "debug"
    if toolchain_flags.strip == "always":
      strip = True
    elif toolchain_flags.strip == "sometimes":
      strip = not debug
  return struct(
      static = _ternary(
          getattr(ctx.attr, "static", None),
          "static" in ctx.features,
      ),
      race = _ternary(
          getattr(ctx.attr, "race", None),
          "race" in ctx.features,
      ),
      msan = _ternary(
          getattr(ctx.attr, "msan", None),
          "msan" in ctx.features,
      ),
      pure = _ternary(
          getattr(ctx.attr, "pure", None),
          force_pure,
          "pure" in ctx.features,
      ),
      link = LINKMODE_NORMAL,
      debug = debug,
      strip = strip,
  )

def _all_modes():
  modes = []
  for static in [False, True]:
    for race in [False, True]:
      for msan in [False, True]:
        for pure in [False, True]:
          for debug in [False, True]:
            for strip in [True, False]:
              # Skip all invalid combinations
              if strip and debug: continue
              if race and pure: continue
              if static and not pure: continue
              # Skip some combinations that don't make much sense
              if msan and pure: continue
              if msan and race: continue
              if msan and not debug: continue
              # Skip some combinations just to reduce the amount, it's too expensive to do them all
              if msan: continue
              if static: continue
              # Add the resulting mode to the list
              modes.append(struct(
                static = static,
                race = race,
                msan = msan,
                pure = pure,
                link = LINKMODE_NORMAL,
                debug = debug,
                strip = strip,
              ))
  return modes

ALL_MODES = _all_modes()