# mini_tanks

A simple game inspired by Battle City. Built with [Ebitengine](https://github.com/hajimehoshi/ebiten).    
Fight enemy tanks and defend your base.

## How to Run

```bash
git clone https://github.com/DTLP/mini_tanks.git
cd mini_tanks
go run ./cmd/mini_tanks.go
```

In case if you're missing any of the dependencies:  
**For Ubuntu/Debian:**
```
sudo apt-get install libx11-dev libxrandr-dev libgl1-mesa-dev libxcursor-dev \
  libxinerama-dev libxi-dev libxxf86vm-dev
```
**For Fedora:**
```
sudo dnf install libX11-devel libXrandr-devel mesa-libGL-devel \
  libXcursor-devel libXinerama-devel libXi-devel libXxf86vm-devel
```
**For Arch Linux:**
```
sudo pacman -S libx11 libxrandr libglvnd libxcursor libxinerama libxi libxxf86vm
```

![Preview](https://github.com/DTLP/mini_tanks/raw/main/preview.gif)
