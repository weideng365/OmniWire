<template>
  <div class="guide-page">
    <el-page-header @back="$router.back()" title="返回" content="OpenVPN 使用说明" style="margin-bottom: 24px" />

    <el-card style="margin-bottom: 20px">
      <template #header><b>一、Linux 服务端安装 OpenVPN</b></template>
      <el-steps direction="vertical" :active="99">
        <el-step title="安装 OpenVPN">
          <template #description>
            <pre class="code">## Debian/Ubuntu
apt update && apt install -y openvpn curl

## CentOS/RHEL
yum install -y epel-release && yum install -y openvpn curl

## Alpine (Docker)
apk add --no-cache openvpn curl</pre>
          </template>
        </el-step>
        <el-step title="启用 IP 转发">
          <template #description>
            <pre class="code">echo "net.ipv4.ip_forward=1" >> /etc/sysctl.conf
sysctl -p</pre>
          </template>
        </el-step>
        <el-step title="启动 OmniWire 并开启 OpenVPN 服务">
          <template #description>
            <p>在 OmniWire 页面点击「启动服务」，系统会自动生成证书和配置文件。</p>
          </template>
        </el-step>
      </el-steps>
    </el-card>

    <el-card style="margin-bottom: 20px">
      <template #header><b>二、Windows 客户端安装</b></template>
      <ol class="guide-list">
        <li>下载 <a href="https://openvpn.net/community-downloads/" target="_blank">OpenVPN GUI</a> 并安装</li>
        <li>在 OmniWire 用户管理页面创建用户，点击下载按钮获取 <code>.ovpn</code> 配置文件</li>
        <li>将 <code>.ovpn</code> 文件放入 <code>C:\Program Files\OpenVPN\config\</code></li>
        <li>右键系统托盘 OpenVPN 图标 → 连接，输入用户名和密码</li>
      </ol>
    </el-card>

    <el-card style="margin-bottom: 20px">
      <template #header><b>三、Linux/macOS 客户端</b></template>
      <pre class="code">## 安装客户端
apt install -y openvpn   # Ubuntu
brew install openvpn     # macOS

## 连接（交互输入用户名密码）
sudo openvpn --config your-username.ovpn</pre>
    </el-card>

    <el-card style="margin-bottom: 20px">
      <template #header><b>四、Android / iOS</b></template>
      <ol class="guide-list">
        <li>安装 <b>OpenVPN Connect</b>（App Store / Google Play）</li>
        <li>下载 <code>.ovpn</code> 配置文件，通过文件管理器或邮件导入 App</li>
        <li>输入用户名和密码连接</li>
      </ol>
    </el-card>

    <el-card>
      <template #header><b>五、分流说明</b></template>
      <p>在「服务配置」中选择<b>分流</b>模式，填写需要走 VPN 的 IP 段（CIDR 格式，逗号分隔），例如：</p>
      <pre class="code">10.0.0.0/8, 192.168.1.0/24, 172.16.0.0/12</pre>
      <p>只有这些 IP 段的流量会经过 VPN，其余流量走本地网络，不影响正常上网。</p>
      <p>修改配置后需<b>重启 OpenVPN 服务</b>才能生效，并重新下载客户端配置文件。</p>
    </el-card>
  </div>
</template>

<script setup></script>

<style scoped>
.guide-page { padding: 20px; max-width: 860px; }
.code {
  background: var(--el-fill-color-light);
  border-radius: 6px;
  padding: 12px 16px;
  font-size: 13px;
  line-height: 1.6;
  white-space: pre-wrap;
  margin: 8px 0;
}
.guide-list { padding-left: 20px; line-height: 2; }
.guide-list code { background: var(--el-fill-color); padding: 2px 6px; border-radius: 4px; }
a { color: var(--el-color-primary); }
</style>
