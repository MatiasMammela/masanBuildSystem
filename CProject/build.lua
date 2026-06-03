mbs = require("mbs")

local project = mbs.project("myProject")
local sources = mbs.glob_files("src/*")
local headers = mbs.glob_dirs("headers")
local libraries = mbs.glob_libraries("lz")

mbs.packages(project,libraries)
mbs.sources(project,sources)
mbs.headers(project,headers)
mbs.build(project)
mbs.debug(project)

