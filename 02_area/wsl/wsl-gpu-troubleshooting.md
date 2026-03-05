---
title: WSL GPU troubleshooting
---

1. Check direcx device

This is the communication channel between `WSL` and `DirectX`.

```sh
ls /dev/dxg
```

- If present: WSL has GPU passthrough support enabled at the kernel level
- If missing: GPU support is completely disabled or not available

2. Check what is rendering OpenGL

Using `glxinfo` we can query the `OpenGL/GLX` implementation.

```sh
glxinfo | grep -E "OpenGL"

# OpenGL vendor string: Mesa
# OpenGL renderer string: llvmpipe (LLVM 20.1.2, 256 bits
```

If it shows `llvmpipe` instead of something like `radeonsi` then there is problem.

llvmpipe = Software renderer (CPU-based)

- "llvm" = Uses LLVM compiler to JIT-compile shader code
- "pipe" = Part of Mesa's Gallium3D driver architecture
- It's a fallback when GPU acceleration fails
- Uses your CPU cores to do what the GPU should do

3. Check Vulkan support

Vulkan = Modern graphics/compute API (alternative to OpenGL).

```sh
vulkaninfo --summary
```

If it is installed it will show something like

```sh
GPU0:
    deviceName     = AMD Radeon RX 7700 XT
    deviceType     = PHYSICAL_DEVICE_TYPE_DISCRETE_GPU

```

4. Check Mesa/D3D12 drivers

```sh
dpkg -l | grep -E "mesa|d3d12"

# d3d12_dri.so -> libdril_dri.so
```

- Mesa = Open-source graphics driver stack for Linux
- D3D12 = DirectX 12 (Windows graphics API)
- WSL uses a special "d3d12" Mesa driver that translates OpenGL/Vulkan → DirectX → Your Windows GPU

If it shows the D3D12 translation layer IS installed.

This is the driver that should make GPU work in WSL. If it is installed but now working it means there is a problem in the communication between this driver and Windows

5. Check kernel errors

```sh
dmesg | grep -i "dxg\|d3d"
```
