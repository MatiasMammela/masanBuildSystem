mbs = require("mbs")

local project = mbs.project("myProject")
local sources = mbs.glob_files("src/*")
local headers = mbs.glob_dirs("headers")
local packages = mbs.glob_packages("sdl2")
local libraries = mbs.glob_libraries("SDL2")

mbs.packages(project,libraries)
mbs.sources(project,sources)
mbs.headers(project,headers)
mbs.build(project)
mbs.debug(project)

