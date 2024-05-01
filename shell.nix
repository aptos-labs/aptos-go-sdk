{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  nativeBuildInputs = with pkgs; [
    go
    go-outline
    gopls
    gopkgs
    go-tools
    delve
  ];
}