{ packages ? <nixpkgs>, modules ? [ ] }:
let
  pkgs = if builtins.isAttrs packages then packages else (import packages { });

  cfg = (import "${toString pkgs.path}/nixos/lib/eval-config.nix" {
    inherit pkgs modules;
  }).config.kubernixos;
in { inherit (cfg) build eval; }
