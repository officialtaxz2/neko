export const logout = '退出登录'
export const unsupported = '该浏览器不支持 WebRTC'
export const admin_loggedin = '您已以管理员身份登录'
export const you = '您'
export const somebody = '某人'
export const send_a_message = '发送消息'

export const side = {
  chat: '聊天',
  files: '文件',
  settings: '设置',
  users: '用户',
}

export const connect = {
  login_title: '请登录',
  invitation_title: '您已被邀请加入此房间',
  displayname: '输入您的显示名称',
  password: '密码',
  connect: '连接',
  error: '登录错误',
  empty_displayname: '显示名称不能为空。',
}

export const context = {
  ignore: '忽略',
  unignore: '取消忽略',
  mute: '静音',
  unmute: '取消静音',
  release: '强制释放控制',
  take: '强制获取控制',
  give: '给予控制',
  kick: '踢出',
  ban: '封禁IP',
  confirm: {
    kick_title: '踢出 {name}？',
    kick_text: '您确定要踢出 {name} 吗？',
    ban_title: '封禁 {name}？',
    ban_text: '您想封禁 {name} 吗？需要重启服务器才能撤销。',
    mute_title: '静音 {name}？',
    mute_text: '您确定要静音 {name} 吗？',
    unmute_title: '取消静音 {name}？',
    unmute_text: '您想取消静音 {name} 吗？',
    button_yes: '是',
    button_cancel: '取消',
  },
}

export const controls = {
  release: '释放控制',
  request: '请求控制',
  lock: '锁定控制',
  unlock: '解锁控制',
  has: '您拥有控制权',
  hasnot: '您没有控制权',
  mic_on: '启用麦克风',
  mic_off: '禁用麦克风',
  mic_error: '麦克风错误',
}

export const locks = {
  control: {
    lock: '锁定控制（针对用户）',
    unlock: '解锁控制（针对用户）',
    locked: '控制已锁定（针对用户）',
    unlocked: '控制已解锁（针对用户）',
    notif_locked: '为用户锁定了控制',
    notif_unlocked: '为用户解锁了控制',
  },
  login: {
    lock: '锁定房间（针对用户）',
    unlock: '解锁房间（针对用户）',
    locked: '房间已锁定（针对用户）',
    unlocked: '房间已解锁（针对用户）',
    notif_locked: '锁定了房间',
    notif_unlocked: '解锁了房间',
  },
  file_transfer: {
    lock: '锁定文件传输（针对用户）',
    unlock: '解锁文件传输（针对用户）',
    locked: '文件传输已锁定（针对用户）',
    unlocked: '文件传输已解锁（针对用户）',
    notif_locked: '锁定了文件传输',
    notif_unlocked: '解锁了文件传输',
  },
}

export const setting = {
  scroll: '滚动灵敏度',
  scroll_invert: '反转滚动',
  trackpad_mode: '触控板模式',
  trackpad_mode_description: '像触控板一样相对移动光标',
  trackpad_mode_mobile_hint: '双指滚动，点击单击，长按右键',
  mobile: '移动端',
  desktop_only_inactive: '仅触摸',
  autoplay: '自动播放视频',
  ignore_emotes: '忽略表情',
  chat_sound: '播放聊天声音',
  keyboard_layout: '键盘布局',
  broadcast_title: '直播',
}

export const connection = {
  logged_out: '您已退出登录。',
  reconnecting: '重新连接中...',
  connected: '已连接',
  disconnected: '已断开',
  kicked: '您已被移出此房间。',
  button_confirm: '确定',
}

export const notifications = {
  connected: '{name} 已连接',
  disconnected: '{name} 已断开',
  controls_taken: '{name} 获取了控制权',
  controls_taken_force: '强制获取了控制权',
  controls_taken_steal: '从 {name} 处获取了控制权',
  controls_released: '{name} 释放了控制权',
  controls_released_force: '强制释放了控制权',
  controls_released_steal: '从 {name} 处释放了控制权',
  controls_given: '将控制权给予了 {name}',
  controls_has: '{name} 拥有控制权',
  controls_has_alt: '但我告知了对方您想要它',
  controls_requesting: '{name} 正在请求控制权',
  resolution: '将分辨率更改为 {width}x{height}@{rate}',
  banned: '封禁了 {name}',
  kicked: '踢出了 {name}',
  muted: '静音了 {name}',
  unmuted: '取消了 {name} 的静音',
}

export const files = {
  downloads: '下载',
  uploads: '上传',
  upload_here: '点击或拖拽文件到此处上传',
}
