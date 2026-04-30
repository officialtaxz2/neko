export const logout = 'Odhlásiť sa'
export const unsupported = 'Tento prehliačač nepodporuje WebRTC'
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
  login_title: 'Prihláste sa',
  invitation_title: 'Boli ste pozvaný do tejto miestnosti',
  displayname: 'Zadajte svoje zobrazovacie meno',
  password: 'Heslo',
  connect: 'Pripojiť',
  error: 'Chyba prihlásenia',
  empty_displayname: 'Zobrazovacie meno nemôže byť prázdne.',
}

export const context = {
  ignore: 'Ignorovať',
  unignore: 'Prestať ignorovať',
  mute: 'Stíšiť',
  unmute: 'Zrušiť stíšenie',
  release: 'Vynútiť uvoľnenie ovládania',
  take: 'Vynútiť prevzatie ovládania',
  give: 'Dať ovládanie',
  kick: 'Vyhoдиť',
  ban: 'Zablokovať IP',
  confirm: {
    kick_title: 'Vyhoдиť {name}?',
    kick_text: 'Naozaj chcete vyhoдиť {name}?',
    ban_title: 'Zablokovať {name}?',
    ban_text: 'Chcete zablokovať {name}? Na zrušenie budete musieť reštartovať server.',
    mute_title: 'Stíšiť {name}?',
    mute_text: 'Naozaj chcete stíšiť {name}?',
    unmute_title: 'Zrušiť stíšenie {name}?',
    unmute_text: 'Chcete zrušiť stíšenie {name}?',
    button_yes: 'Áno',
    button_cancel: 'Zrušiť',
  },
}

export const controls = {
  release: 'Uvoľniť ovládanie',
  request: 'Požиадаť o ovládanie',
  lock: 'Zamknúť ovládanie',
  unlock: 'Odomknúť ovládanie',
  has: 'Máte ovládanie',
  hasnot: 'Nemáte ovládanie',
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
    lock: 'Zamknúť miestnostť (pre používateľov)',
    unlock: 'Odomknúť miestnostť (pre používateľov)',
    locked: 'Miestnostť zamknutá (pre používateľov)',
    unlocked: 'Miestnostť odomknutá (pre používateľov)',
    notif_locked: 'zamkol miestnostť',
    notif_unlocked: 'odomkol miestnostť',
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
  controls_taken_force: 'násilne prevzal ovládanie',
  controls_taken_steal: 'prevzal ovládanie od {name}',
  controls_released: '{name} uvoľnil ovládanie',
  controls_released_force: 'násilne uvoľnil ovládanie',
  controls_released_steal: 'uvoľnil ovládanie od {name}',
  controls_given: 'dal ovládanie {name}',
  controls_has: '{name} má ovládanie',
  controls_has_alt: 'Ale dal som tej osobe vedieť, že ste to chceli',
  controls_requesting: '{name} žiada o ovládanie',
  resolution: 'zmenil rozlíšenie na {width}x{height}@{rate}',
  banned: 'zablokoval {name}',
  kicked: 'vyhodil {name}',
  muted: 'stíšil {name}',
  unmuted: 'zrušil stíšenie {name}',
}

export const files = {
  downloads: 'Stiahnuté',
  uploads: 'Nahraté',
  upload_here: 'Kliknite alebo potiáhnite súbory sem na nahranie',
}
