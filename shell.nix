{ pkgs ? import (fetchTarball "https://github.com/NixOS/nixpkgs/archive/nixos-26.05.tar.gz") { } }:

pkgs.mkShell {
  packages = with pkgs; [
    go_1_25
    gnumake
    pkg-config
    vips
  ];

  CGO_ENABLED = "1";
}
