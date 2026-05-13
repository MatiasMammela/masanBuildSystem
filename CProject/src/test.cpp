#include "test.h"

test::test()
{
    // ncurses
    initscr();
    printw("Hello from ncurses!");
    refresh();
    getch();
    endwin();


    uLong srcLen = strlen("hello world") + 1;
    uLong destLen = compressBound(srcLen);
    Bytef* dest = new Bytef[destLen];
    compress(dest, &destLen, (const Bytef*)"hello world", srcLen);
    
    // SDL2
    if (SDL_Init(SDL_INIT_VIDEO) != 0)
    {
        std::cerr << "SDL_Init Error: " << SDL_GetError() << std::endl;
        return;
    }
    SDL_Quit();
}

test::~test() {}
