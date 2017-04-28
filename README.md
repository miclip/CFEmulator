# cf-emulator

To run:

  1. Create a directory ```c:\containerizer```
  2. Clone and build Containerizer and place build output in ```c:\containerizer```. [`windows garden`](https://github.com/cloudfoundry/garden-windows)
  3. Clone and build hwc.exe and copy to ```c:\containerizer```. [`hwc`](https://github.com/cloudfoundry-incubator/hwc-buildpack)
  4. Enable disk quotas on C DRIVE.
  5. Ensure windows firewall is enabled.
  6. Run cf-emulator: ```cf-emulator --path c:\Projects\MyApplication```
