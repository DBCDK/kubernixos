{ pkgs ? (import <nixpkgs> {}) }:
# PIN dis 
with pkgs;
let
  dep2nix = callPackage ./nix-packaging/dep2nix {};
  kubernixos = pkgs.callPackage ./default.nix {};
  in
  # Change to mkShell once that hits stable!
  pkgs.mkShell {
    name = "kubernixos-build-env";

  inputsFrom = [
    kubernixos
  ];
  
}
