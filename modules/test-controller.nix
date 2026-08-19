{ pkgs, ... }:

{
  services.xserver.enable = true;
  services.xserver.displayManager.startx.enable = true;
  virtualisation.graphics = true;
  virtualisation.qemu.options = [ "-vga virtio" ];
}
