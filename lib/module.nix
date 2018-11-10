{ config, pkgs, lib, ... }:

let
  cfg = config.kubernixos;

  kubeval = pkgs.callPackage ./kubeval {};
  schemas = pkgs.callPackage ./schemas {};

  assertion = with pkgs; with builtins; name: item:
  let
    validate = runCommand "validate-${name}" {} ''
      mkdir -p $out
      echo '${toJSON item}' | \
        ${kubeval}/bin/kubeval --strict -v ${cfg.version} \
        --filename=${name} \
        --schema-location=file://${schemas}

      echo -n "true" >$out/result
  '';
  in
  {
    assertion = (readFile "${validate}/result") == "true";
    message = "${name} has validation errors!";
  };

in
{

  options.kubernixos = with lib; with lib.types; {

    manifests = mkOption {
      type = attrsOf attrs;
      default = {};
      description = "Attribute set of kubernetes manifests.";
    };

    version = mkOption {
      type = enum [ "1.11.0" "1.11.1" ];
      default = "1.11.1";
      description = "Kubernetes version to apply manifest validation against.";
    };

    assertions = mkOption {
      type = listOf attrs;
      default = [];
      description = "Kubernixos assertions to eval before generating k8s config.";
    };

  };

  config.kubernixos.assertions = lib.mapAttrsToList assertion cfg.manifests;
  # Pass kubernixos assertions to config.assertions in order to get them eval'ed when
  # building NixOS systems.
  config.assertions = config.kubernixos.assertions;
}
