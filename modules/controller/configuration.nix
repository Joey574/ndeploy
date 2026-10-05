{ pkgs, ndeploy, ... }:

let
  home = "/home/controller"
  configDir = "${home}/nixos"
in
{
  imports = [
    ./ssh.nix
  ];

  nix.nixPath = [ "nixpkgs=${pkgs.path}" ];
  nix.settings.trusted-users = [
    "root"
    "controller"
  ];

  services.xserver.enable = true;
  services.xserver.windowManager.i3.enable = true;

  users.users.controller = {
    isNormalUser = true;
    password = "";
    extraGroups = [
      "video"
      "input"
      "wheel"
    ];
  };

  systemd.tmpfiles.rules = [
    "d ${home}/.cache 0755 controller users -"
    "d ${home}/.config 0755 controller users -"

    "d ${configDir} 0755 controller users -"
    "C ${configDir}/controller 0755 controller users - ${./.}"
    "C ${configDir}/node 0755 controller users - ${../nodet1}"
    "Z ${configDir} - controller users -"
  ];

  services.displayManager.sddm.enable = true;
  services.desktopManager.plasma6.enable = true;
  services.displayManager.defaultSession = "plasma";

  services.displayManager.autoLogin = {
    enable = true;
    user = "controller";
  };

  virtualisation = {
    graphics = true;
    qemu.options = [
      "-vga virtio"
      "-display gtk"
      "-device usb-tablet"
    ];

    cores = 6;
    memorySize = 16384;
    resolution = {
      x = 1920;
      y = 1080;
    };
  };

  environment.systemPackages = with pkgs; [
    mesa
  ] ++ lib.optional (ndeploy != null) ndeploy;
  system.stateVersion = "26.05";
}
