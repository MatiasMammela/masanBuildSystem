mbs = require("mbs")
local project = mbs.project("myproject")
local sources = mbs.glob_files("src/*")
local packages= mbs.glob_packages("sdl2")
local headers = mbs.glob_dirs("headers")

mbs.sources(project,sources)
mbs.headers(project,headers)

mbs.packages(project,packages)

mbs.lflags(project, "--static")
mbs.build(project)
mbs.debug(project)