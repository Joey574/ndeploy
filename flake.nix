{
  description = "ndeploy fleet manager";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    { self, nixpkgs, ... }:
    let
      system = "x86_64-linux";
      pkgs = import nixpkgs { inherit system; };
      lib = nixpkgs.lib;
    in
    {
      packages.${system}.ndeploy = pkgs.buildGoModule {
        pname = "ndeploy";
        version = "0-unstable";
        src = ./.;

        proxyVendor = true;
        vendorHash = "sha256-3ZPNaMmINK+1ouh/jRephF/H5hDPtctGtREvq33Rmxo=";
        #vendorHash = lib.fakeHash;

        nativeBuildInputs = with pkgs; [
          pkg-config
          copyDesktopItems
        ];

        buildInputs = with pkgs; [
          libGL
          libX11
          libxcursor
          libxrandr
          libxinerama
          libxi
          libxxf86vm
          wayland
          wayland-protocols
          libxkbcommon
        ];

        desktopItems = [
          (pkgs.makeDesktopItem {
            name = "ndeploy";
            desktopName = "ndeploy";
            genericName = "NixOS Fleet Manager";
            comment = "Deploy and manage a fleet of NixOS nodes";
            exec = "ndeploy";
            icon = "ndeploy";
            categories = [
              "System"
              "Utility"
            ];
            startupWMClass = "ndeploy";
            terminal = false;
          })
        ];

        postInstall = ''
          install -Dm644 assets/icon.svg \
              $out/share/icons/hicolor/scalable/apps/ndeploy.svg
        '';

        meta.mainProgram = "ndeploy";
      };

      checks.${system}.fleet-test = pkgs.testers.runNixOSTest {
        name = "ndeploy-test";

        nodes = {
          controller = { pkgs, ... }: {
            imports = [ ./modules/controller/configuration.nix ];
            environment.systemPackages = [ self.packages.${system}.ndeploy ];
            _module.args.ndeploy = self.packages.${system}.ndeploy;
          };

          node1 = { ... }: { imports = [ ./modules/nodet1/configuration.nix ]; };
          node2 = { ... }: { imports = [ ./modules/nodet1/configuration.nix ]; };
        };

        testScript = ''
          start_all()
          node1.wait_for_unit("sshd.service")
          node2.wait_for_unit("sshd.service")
          controller.wait_for_unit("multi-user.target")
          controller.wait_for_unit("display-manager.service")
          controller.sleep(3)
        '';
      };
    };
}
