{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    go
    gopls
    pkg-config

    libX11
    libXcursor
    libXrandr
    libXinerama
    libXi
    libXxf86vm

    libGL
    wayland
    libxkbcommon
  ];

  shellHook = ''
    export LD_LIBRARY_PATH=${pkgs.lib.makeLibraryPath (with pkgs; [
      libGL
      wayland
      libxkbcommon
      libX11
      libXcursor
      libXrandr
      libXi
      libXxf86vm
      libXinerama
    ])}:$LD_LIBRARY_PATH
  '';
}
