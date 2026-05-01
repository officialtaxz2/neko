export const logout = 'Odhlásiť sa'
export const unsupported = 'Tento webový prehliadač nepodporuje WebRTC'
export const admin_loggedin = 'Ste prihlásený ako správca'
export const you = 'Vy'
export const somebody = 'Niekto'
export const send_a_message = 'Odoslať správu'

export const side = {
  chat: 'Chat',
  files: 'Súbory',
  settings: 'Nastavenia',
  users: 'Používatelia',
}

export const connect = {
  login_title: 'Prihláste sa prosím',
  invitation_title: 'Boli ste pozvaný do tejto miestnosti',
  displayname: 'Zadajte svoje zobrazované meno',
  password: 'Heslo',
  connect: 'Pripojiť',
  error: 'Chyba prihlásenia',
  empty_displayname: 'Zobrazované meno nemôže byť prázdne.',
}

export const context = {
  ignore: 'Ignorovať',
  unignore: 'Prestať ignorovať',
  mute: 'Stlmiť',
  unmute: 'Zrušiť stlmenie',
  release: 'Vynútiť uvoľnenie ovládania',
  take: 'Vynútiť prevzatie ovládania',
  give: 'Odovzdať ovládanie',
  kick: 'Vyhodiť',
  ban: 'Zablokovať IP',
  confirm: {
    kick_title: 'Vyhodiť {name}?',
    kick_text: 'Ste si istý, že chcete vyhodiť {name}?',
    ban_title: 'Zablokovať {name}?',
    ban_text: 'Chcete zablokovať {name}? Budete musieť reštartovať server, aby ste to vrátili.',
    mute_title: 'Stlmiť {name}?',
    mute_text: 'Ste si istý, že chcete stlmiť {name}?',
    unmute_title: 'Zrušiť stlmenie {name}?',
    unmute_text: 'Chcete zrušiť stlmenie {name}?',
    button_yes: 'Áno',
    button_cancel: 'Zrušiť',
  },
}

export const controls = {
  release: 'Uvoľniť ovládanie',
  request: 'Požiadať o ovládanie',
  lock: 'Zamknúť ovládanie',
  unlock: 'Odomknúť ovládanie',
  has: 'Máte ovládanie',
  hasnot: 'Nemáte ovládanie',
  mic_on: 'Zapnúť mikrofón',
  mic_off: 'Vypnúť mikrofón',
  mic_error: 'Chyba mikrofónu',
}

export const locks = {
  control: {
    lock: 'Zamknúť ovládanie (pre používateľov)',
    unlock: 'Odomknúť ovládanie (pre používateľov)',
    locked: 'Ovládanie zamknuté (pre používateľov)',
    unlocked: 'Ovládanie odomknuté (pre používateľov)',
    notif_locked: 'zamkol ovládanie pre používateľov',
    notif_unlocked: 'odomkol ovládanie pre používateľov',
  },
  login: {
    lock: 'Zamknúť miestnosť (pre používateľov)',
    unlock: 'Odomknúť miestnosť (pre používateľov)',
    locked: 'Miestnosť zamknutá (pre používateľov)',
    unlocked: 'Miestnosť odomknutá (pre používateľov)',
    notif_locked: 'zamkol miestnosť',
    notif_unlocked: 'odomkol miestnosť',
  },
  file_transfer: {
    lock: 'Zamknúť prenos súborov (pre používateľov)',
    unlock: 'Odomknúť prenos súborov (pre používateľov)',
    locked: 'Prenos súborov zamknutý (pre používateľov)',
    unlocked: 'Prenos súborov odomknutý (pre používateľov)',
    notif_locked: 'zamkol prenos súborov',
    notif_unlocked: 'odomkol prenos súborov',
  },
}

export const setting = {
  scroll: 'Citlivosť posúvania',
  scroll_invert: 'Invertovať posúvanie',
  trackpad_mode: 'Režim touchpadu',
  trackpad_mode_description: 'Pohybovať kurzorom relatívne ako touchpad',
  trackpad_mode_mobile_hint: 'Dva prsty na rolovanie, klepnutie na klik, podržanie na pravý klik',
  mobile: 'Mobil',
  desktop_only_inactive: 'Len dotyk',
  autoplay: 'Automatické prehrávanie videa',
  ignore_emotes: 'Ignorovať emotikony',
  chat_sound: 'Prehrávať zvuk chatu',
  keyboard_layout: 'Rozloženie klávesnice',
  broadcast_title: 'Živé vysielanie',
}

export const connection = {
  logged_out: 'Boli ste odhlásený.',
  reconnecting: 'Opätovné pripájanie...',
  connected: 'Pripojený',
  disconnected: 'Odpojený',
  kicked: 'Boli ste odstránený z tejto miestnosti.',
  button_confirm: 'OK',
}

export const notifications = {
  connected: '{name} sa pripojil',
  disconnected: '{name} sa odpojil',
  controls_taken: '{name} prevzal ovládanie',
  controls_taken_force: 'vynútene prevzal ovládanie',
  controls_taken_steal: 'prevzal ovládanie od {name}',
  controls_released: '{name} uvoľnil ovládanie',
  controls_released_force: 'vynútene uvoľnil ovládanie',
  controls_released_steal: 'uvoľnil ovládanie od {name}',
  controls_given: 'odovzdal ovládanie {name}',
  controls_has: '{name} má ovládanie',
  controls_has_alt: 'Ale informoval som danú osobu, že ste to chceli',
  controls_requesting: '{name} žiada o ovládanie',
  resolution: 'zmenil rozlíšenie na {width}x{height}@{rate}',
  banned: 'zablokoval {name}',
  kicked: 'vyhodil {name}',
  muted: 'stlmil {name}',
  unmuted: 'zrušil stlmenie {name}',
}

export const files = {
  downloads: 'Stiahnuté',
  uploads: 'Nahrané',
  upload_here: 'Kliknite alebo presuňte súbory sem na nahranie',
}
