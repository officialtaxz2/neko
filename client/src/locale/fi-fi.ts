export const logout = 'Kirjaudu ulos'
export const unsupported = 'Tämä selain ei tue WebRTC:tä'
export const admin_loggedin = 'Olet kirjautunut sisään ylläpitäjänä'
export const you = 'Sinä'
export const somebody = 'Joku'
export const send_a_message = 'Lähetä viesti'

export const side = {
  chat: 'Chat',
  files: 'Tiedostot',
  settings: 'Asetukset',
  users: 'Käyttäjät',
}

export const connect = {
  login_title: 'Ole hyvä ja kirjaudu sisään',
  invitation_title: 'Sinut on kutsuttu tähän huoneeseen',
  displayname: 'Kirjoita näyttönimesi',
  password: 'Salasana',
  connect: 'Yhdistä',
  error: 'Kirjautumisvirhe',
  empty_displayname: 'Näyttönimi ei voi olla tyhjä.',
}

export const context = {
  ignore: 'Ohita',
  unignore: 'Lopeta ohittaminen',
  mute: 'Mykistä',
  unmute: 'Poista mykistys',
  release: 'Pakota vapautus',
  take: 'Pakota haltuunotto',
  give: 'Anna hallinta',
  kick: 'Poista',
  ban: 'Bannaa IP',
  confirm: {
    kick_title: 'Poistetaanko {name}?',
    kick_text: 'Oletko varma, että haluat poistaa {name}?',
    ban_title: 'Bannataan {name}?',
    ban_text: 'Haluatko bannata {name}? Sinun täytyy käynnistää palvelin uudelleen peruuttaaksesi tämän.',
    mute_title: 'Mykistetäänkö {name}?',
    mute_text: 'Oletko varma, että haluat mykistää {name}?',
    unmute_title: 'Poistetaanko {name}:n mykistys?',
    unmute_text: 'Haluatko poistaa {name}:n mykistyksen?',
    button_yes: 'Kyllä',
    button_cancel: 'Peruuta',
  },
}

export const controls = {
  release: 'Vapauta hallinta',
  request: 'Pyydä hallintaa',
  lock: 'Lukitse hallinta',
  unlock: 'Avaa hallinta',
  has: 'Sinulla on hallinta',
  hasnot: 'Sinulla ei ole hallintaa',
  mic_on: 'Ota mikrofoni käyttöön',
  mic_off: 'Poista mikrofoni käytöstä',
  mic_error: 'Mikrofonin virhe',
}

export const locks = {
  control: {
    lock: 'Lukitse hallinta (käyttäjille)',
    unlock: 'Avaa hallinta (käyttäjille)',
    locked: 'Hallinta lukittu (käyttäjille)',
    unlocked: 'Hallinta avattu (käyttäjille)',
    notif_locked: 'lukitsi hallinnan käyttäjiltä',
    notif_unlocked: 'avasi hallinnan käyttäjille',
  },
  login: {
    lock: 'Lukitse huone (käyttäjille)',
    unlock: 'Avaa huone (käyttäjille)',
    locked: 'Huone lukittu (käyttäjille)',
    unlocked: 'Huone avattu (käyttäjille)',
    notif_locked: 'lukitsi huoneen',
    notif_unlocked: 'avasi huoneen',
  },
  file_transfer: {
    lock: 'Lukitse tiedostonsiirto (käyttäjille)',
    unlock: 'Avaa tiedostonsiirto (käyttäjille)',
    locked: 'Tiedostonsiirto lukittu (käyttäjille)',
    unlocked: 'Tiedostonsiirto avattu (käyttäjille)',
    notif_locked: 'lukitsi tiedostonsiirron',
    notif_unlocked: 'avasi tiedostonsiirron',
  },
}

export const setting = {
  scroll: 'Scrollin herkkyys',
  scroll_invert: 'Käänteinen Scroll',
  trackpad_mode: 'Kosketuslevy-tila',
  trackpad_mode_description: 'Siirrä kursoria suhteellisesti kuten kosketuslevyllä',
  trackpad_mode_mobile_hint: 'Kaksi sormea vierittämiseen, napauta klikkaamiseen, pidä pohjassa oikealle klikkaukselle',
  mobile: 'Mobiili',
  desktop_only_inactive: 'Vain kosketusnäyttö',
  autoplay: 'Automaattisesti toista video',
  ignore_emotes: 'Estä emojit',
  chat_sound: 'Soita viesti ääni',
  keyboard_layout: 'Näppäimistöasettelu',
  broadcast_title: 'Suora Lähetys',
}

export const connection = {
  logged_out: 'Sinut on kirjattu ulos.',
  reconnecting: 'Yhteyttä yritetään palauttaa...',
  connected: 'Yhdistetty',
  disconnected: 'Katkaistu yhteys',
  kicked: 'Sinut on poistettu huoneesta.',
  button_confirm: 'OK',
}

export const notifications = {
  connected: '{name} liittyi',
  disconnected: '{name} poistui',
  controls_taken: '{name} otti hallinnan',
  controls_taken_force: 'otti hallinnan pakolla',
  controls_taken_steal: 'otti hallinnan {name}:lta',
  controls_released: '{name} vapautti hallinnan',
  controls_released_force: 'vapautti hallinnan pakolla',
  controls_released_steal: 'vapautti hallinnan {name}:lta',
  controls_given: 'antoi hallinnan {name}:lle',
  controls_has: '{name} on hallinnassa',
  controls_has_alt: 'Mutta kerroin henkilölle, että halusit sen',
  controls_requesting: '{name} pyytää hallintaa',
  resolution: 'muutti resoluutioksi {width}x{height}@{rate}',
  banned: 'bannasi {name}',
  kicked: 'poisti {name}',
  muted: 'mykisti {name}',
  unmuted: 'poisti {name}:n mykistyksen',
}

export const files = {
  downloads: 'Lataukset',
  uploads: 'Lähetykset',
  upload_here: 'Napsauta tai vedä tiedostot tähän ladataksesi',
}
