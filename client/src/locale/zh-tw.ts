export const logout = '登出'
export const unsupported = '此瀏覽器不支援 WebRTC'
export const admin_loggedin = '您已以管理員身分登入'
export const you = '您'
export const somebody = '某人'
export const send_a_message = '傳送訊息'

export const side = {
  chat: '聊天',
  files: '檔案',
  settings: '設定',
  users: '使用者',
}

export const connect = {
  login_title: '請登入',
  invitation_title: '您已被邀請加入此房間',
  displayname: '輸入您的顯示名稱',
  password: '密碼',
  connect: '連線',
  error: '登入錯誤',
  empty_displayname: '顯示名稱不能為空。',
}

export const context = {
  ignore: '忽略',
  unignore: '取消忽略',
  mute: '靜音',
  unmute: '取消靜音',
  release: '強制釋放控制',
  take: '強制取得控制',
  give: '給予控制',
  kick: '踢出',
  ban: '封鎖IP',
  confirm: {
    kick_title: '踢出 {name}？',
    kick_text: '您確定要踢出 {name} 嗎？',
    ban_title: '封鎖 {name}？',
    ban_text: '您想封鎖 {name} 嗎？需要重新啟動伺服器才能取消。',
    mute_title: '靜音 {name}？',
    mute_text: '您確定要靜音 {name} 嗎？',
    unmute_title: '取消靜音 {name}？',
    unmute_text: '您想取消靜音 {name} 嗎？',
    button_yes: '是',
    button_cancel: '取消',
  },
}

export const controls = {
  release: '釋放控制',
  request: '請求控制',
  lock: '鎖定控制',
  unlock: '解鎖控制',
  has: '您擁有控制權',
  hasnot: '您沒有控制權',
  mic_on: '啟用麥克風',
  mic_off: '停用麥克風',
  mic_error: '麥克風錯誤',
}

export const locks = {
  control: {
    lock: '鎖定控制（針對使用者）',
    unlock: '解鎖控制（針對使用者）',
    locked: '控制已鎖定（針對使用者）',
    unlocked: '控制已解鎖（針對使用者）',
    notif_locked: '為使用者鎖定了控制',
    notif_unlocked: '為使用者解鎖了控制',
  },
  login: {
    lock: '鎖定房間（針對使用者）',
    unlock: '解鎖房間（針對使用者）',
    locked: '房間已鎖定（針對使用者）',
    unlocked: '房間已解鎖（針對使用者）',
    notif_locked: '鎖定了房間',
    notif_unlocked: '解鎖了房間',
  },
  file_transfer: {
    lock: '鎖定檔案傳輸（針對使用者）',
    unlock: '解鎖檔案傳輸（針對使用者）',
    locked: '檔案傳輸已鎖定（針對使用者）',
    unlocked: '檔案傳輸已解鎖（針對使用者）',
    notif_locked: '鎖定了檔案傳輸',
    notif_unlocked: '解鎖了檔案傳輸',
  },
}

export const setting = {
  scroll: '捲動靈敏度',
  scroll_invert: '反轉捲動',
  trackpad_mode: '觸控板模式',
  trackpad_mode_description: '像觸控板一樣相對移動游標',
  trackpad_mode_mobile_hint: '雙指捲動，點擊單擊，長按右鍵',
  mobile: '行動裝置',
  desktop_only_inactive: '僅觸控',
  autoplay: '自動播放影片',
  ignore_emotes: '忽略表情符號',
  chat_sound: '播放聊天聲音',
  keyboard_layout: '鍵盤配置',
  broadcast_title: '直播',
}

export const connection = {
  logged_out: '您已登出。',
  reconnecting: '重新連接中...',
  connected: '已連接',
  disconnected: '已斷開',
  kicked: '您已被移出此房間。',
  button_confirm: '確定',
}

export const notifications = {
  connected: '{name} 已連接',
  disconnected: '{name} 已斷開',
  controls_taken: '{name} 取得了控制權',
  controls_taken_force: '強制取得了控制權',
  controls_taken_steal: '從 {name} 處取得了控制權',
  controls_released: '{name} 釋放了控制權',
  controls_released_force: '強制釋放了控制權',
  controls_released_steal: '從 {name} 處釋放了控制權',
  controls_given: '將控制權給予了 {name}',
  controls_has: '{name} 擁有控制權',
  controls_has_alt: '但我告知了對方您想要它',
  controls_requesting: '{name} 正在請求控制權',
  resolution: '將解析度更改為 {width}x{height}@{rate}',
  banned: '封鎖了 {name}',
  kicked: '踢出了 {name}',
  muted: '靜音了 {name}',
  unmuted: '取消了 {name} 的靜音',
}

export const files = {
  downloads: '下載',
  uploads: '上傳',
  upload_here: '點擊或拖拽檔案到此處上傳',
}
