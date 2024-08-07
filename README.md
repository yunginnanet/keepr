
<p align="center"><a href="https://tcp.ac/i/jDG9s" target="_blank"><img width="500" src="https://tcp.ac/i/jDG9s"></a></p>
<h1 align="center">keepr</h1>
<p align="center">organize your audio samples.. <i>but don't touch them</i>.</p>


## problem

 * too many audio samples
   * 250 gigs scattered about in different subdirectories
   * moving them would immediately cause chaos in past project files

## solution

 * create folder filled with subfolders that **we populate with symlinks**.
   * use file names, wav data, and parent directory names for hints
   * allows for easy browsing of audio samples from any standard DAW browser by:
     * **key**
     * **tempo**
     * **percussion type**
     * whatever we think of next

keepr is **fast**. _really_ fast. on my system the bottleneck becomes I/O. When reading and writing to a single NVMe drive keepr averages around 700MBp/s disk read and spikes up to nearly 2GBp/s disk read.

## will you ever finish it

do I ever finish anything? idk maybe. it works right now better than the old version (which was a shitty bash script that ran fdfind), so it's lookin good so far.

 - [x] guess tempo by filename
 - [x] separate wave files and midi files
 - [x] validate wave files
 - [x] guess key by filename
 - [x] guess drum type by parent directory
 - [x] create symlinks for all of the above\
 - [x] be stupid dumb fast
 - [ ] verify various theories with wave/midi data
 - [ ] sort MIDI files
 - [x] more taxonomy
 - [ ] unit tests
 - [x] in-app documentation
 - [ ] more to-do items

## recognition

 * [kr/walk](https://github.com/kr/walk)
 * [go-audio/wav](https://github.com/go-audio/wav)
 * [go-music-theory/music-theory](https://github.com/go-music-theory/music-theory)
 * [gomidi/midi](https://github.com/gomidi/)

