 { packages ? <nixpkgs>, modules ? [] }:
let
  pkgs = if builtins.isAttrs packages then packages else (import packages {});

  merge = name: item:
    pkgs.lib.recursiveUpdate item {
      metadata = {
        labels = {
          reconciler = "kubernixos";
        };
      };
    };

in
{
  manifests = with pkgs;
  let
    config = (import "${toString path}/nixos/lib/eval-config.nix" {
     inherit pkgs modules;
    }).config.kubernixos;

    # Assertion validation borrowed from /modules/system/activation/top-level.nix
    failedAssertions = with pkgs.lib; map (x: x.message) (filter (x: !x.assertion) config.assertions);
  in
    if failedAssertions != []
    then throw "\nFailed assertions:\n${concatStringsSep "\n" (map (x: "- ${x}") failedAssertions)}"
    else
    {
      apiVersion = "v1";
      kind = "List";
      items = lib.mapAttrsToList merge config.manifests;
    };
}
