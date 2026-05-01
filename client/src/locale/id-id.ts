export const logout = 'Keluar'
export const unsupported = 'Peramban ini tidak mendukung WebRTC'
export const admin_loggedin = 'Anda masuk sebagai admin'
export const you = 'Anda'
export const somebody = 'Seseorang'
export const send_a_message = 'Kirim pesan'

export const side = {
  chat: 'Obrolan',
  files: 'Berkas',
  settings: 'Pengaturan',
  users: 'Pengguna',
}

export const connect = {
  login_title: 'Silakan Masuk',
  invitation_title: 'Anda telah diundang ke ruangan ini',
  displayname: 'Masukkan nama tampilan Anda',
  password: 'Kata sandi',
  connect: 'Hubungkan',
  error: 'Kesalahan login',
  empty_displayname: 'Nama tampilan tidak boleh kosong.',
}

export const context = {
  ignore: 'Abaikan',
  unignore: 'Berhenti mengabaikan',
  mute: 'Bisukan',
  unmute: 'Aktifkan suara',
  release: 'Paksa Lepas Kontrol',
  take: 'Paksa Ambil Kontrol',
  give: 'Berikan Kontrol',
  kick: 'Tendang',
  ban: 'Blokir IP',
  confirm: {
    kick_title: 'Tendang {name}?',
    kick_text: 'Apakah Anda yakin ingin menendang {name}?',
    ban_title: 'Blokir {name}?',
    ban_text: 'Apakah Anda ingin memblokir {name}? Anda perlu me-restart server untuk membatalkannya.',
    mute_title: 'Bisukan {name}?',
    mute_text: 'Apakah Anda yakin ingin membisukan {name}?',
    unmute_title: 'Aktifkan suara {name}?',
    unmute_text: 'Apakah Anda ingin mengaktifkan suara {name}?',
    button_yes: 'Ya',
    button_cancel: 'Batal',
  },
}

export const controls = {
  release: 'Lepas Kontrol',
  request: 'Minta Kontrol',
  lock: 'Kunci Kontrol',
  unlock: 'Buka Kunci Kontrol',
  has: 'Anda memiliki kontrol',
  hasnot: 'Anda tidak memiliki kontrol',
  mic_on: 'Aktifkan Mikrofon',
  mic_off: 'Nonaktifkan Mikrofon',
  mic_error: 'Kesalahan Mikrofon',
}

export const locks = {
  control: {
    lock: 'Kunci Kontrol (untuk pengguna)',
    unlock: 'Buka Kunci Kontrol (untuk pengguna)',
    locked: 'Kontrol Dikunci (untuk pengguna)',
    unlocked: 'Kontrol Dibuka (untuk pengguna)',
    notif_locked: 'mengunci kontrol untuk pengguna',
    notif_unlocked: 'membuka kontrol untuk pengguna',
  },
  login: {
    lock: 'Kunci Ruangan (untuk pengguna)',
    unlock: 'Buka Ruangan (untuk pengguna)',
    locked: 'Ruangan Dikunci (untuk pengguna)',
    unlocked: 'Ruangan Dibuka (untuk pengguna)',
    notif_locked: 'mengunci ruangan',
    notif_unlocked: 'membuka ruangan',
  },
  file_transfer: {
    lock: 'Kunci Transfer Berkas (untuk pengguna)',
    unlock: 'Buka Transfer Berkas (untuk pengguna)',
    locked: 'Transfer Berkas Dikunci (untuk pengguna)',
    unlocked: 'Transfer Berkas Dibuka (untuk pengguna)',
    notif_locked: 'mengunci transfer berkas',
    notif_unlocked: 'membuka transfer berkas',
  },
}

export const setting = {
  scroll: 'Sensitivitas Gulir',
  scroll_invert: 'Gulir Terbalik',
  trackpad_mode: 'Mode Trackpad',
  trackpad_mode_description: 'Gerakkan kursor secara relatif seperti touchpad',
  trackpad_mode_mobile_hint: 'Dua jari untuk menggulir, ketuk untuk klik, tahan untuk klik kanan',
  mobile: 'Seluler',
  desktop_only_inactive: 'Khusus sentuh',
  autoplay: 'Putar Video Otomatis',
  ignore_emotes: 'Abaikan Emoticon',
  chat_sound: 'Nyalakan Bunyi Obrolan',
  keyboard_layout: 'Tata Letak Papan Tik',
  broadcast_title: 'Siaran Langsung',
}

export const connection = {
  logged_out: 'Anda telah keluar.',
  reconnecting: 'Menyambungkan ulang...',
  connected: 'Tersambung',
  disconnected: 'Terputus',
  kicked: 'Anda telah dikeluarkan dari ruangan ini.',
  button_confirm: 'Oke',
}

export const notifications = {
  connected: '{name} tersambung',
  disconnected: '{name} terputus',
  controls_taken: '{name} mengambil kontrol',
  controls_taken_force: 'mengambil kontrol secara paksa',
  controls_taken_steal: 'mengambil kontrol dari {name}',
  controls_released: '{name} melepas kontrol',
  controls_released_force: 'melepas kontrol secara paksa',
  controls_released_steal: 'melepas kontrol dari {name}',
  controls_given: 'memberikan kontrol kepada {name}',
  controls_has: '{name} memiliki kontrol',
  controls_has_alt: 'Tapi saya memberi tahu orang itu bahwa Anda menginginkannya',
  controls_requesting: '{name} meminta kontrol',
  resolution: 'mengubah resolusi menjadi {width}x{height}@{rate}',
  banned: 'memblokir {name}',
  kicked: 'menendang {name}',
  muted: 'membisukan {name}',
  unmuted: 'mengaktifkan suara {name}',
}

export const files = {
  downloads: 'Unduhan',
  uploads: 'Unggahan',
  upload_here: 'Klik atau seret berkas ke sini untuk mengunggah',
}
