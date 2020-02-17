{ config, pkgs, lib, ... }:

let
  cfg = config.kubernixos;

  assertion = with pkgs; with builtins; name: item:
  let
    validate = runCommand "validate-${name}" {} ''
      mkdir -p $out
      echo '${toJSON item}' | \
        ${pkgs.kubeval}/bin/kubeval --strict -v ${cfg.version} \
        --filename=${name} \
        --schema-location=file://${cfg.schemas}

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

    schemas = mkOption {
      type = package;
      default = pkgs.callPackage ./schemas {};
      description = "k8s jsonschemas package";
    };

    version = mkOption {
      type = str;
      description = "Kubernetes version to apply manifest validation against.";
    };

    assertions = mkOption {
      type = listOf attrs;
      default = [];
      description = "Kubernixos assertions to eval before generating k8s config.";
    };

    server = mkOption {
      type = str;
      description = ''
        Address to the kubernetes apiserver.
        This is a safety to ensure that applying kubeconfigs and manifest conf match.
      '';
    };

  };

  config.kubernixos.assertions = lib.mapAttrsToList assertion cfg.manifests;
  # Pass kubernixos assertions to config.assertions in order to get them eval'ed when
  # building NixOS systems.
  config.assertions = config.kubernixos.assertions;
}
