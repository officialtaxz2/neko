export const logout = 'Se déconnecter'
export const unsupported = 'Ce navigateur ne prend pas en charge WebRTC'
export const admin_loggedin = 'Vous êtes connecté en tant qu\'administrateur'
export const you = 'Vous'
export const somebody = 'Quelqu\'un'
export const send_a_message = 'Envoyer un message'

export const side = {
  chat: 'Chat',
  files: 'Fichiers',
  settings: 'Paramètres',
  users: 'Utilisateurs',
}

export const connect = {
  login_title: 'Veuillez vous connecter',
  invitation_title: 'Vous avez été invité dans cette salle',
  displayname: 'Entrez votre nom d\'affichage',
  password: 'Mot de passe',
  connect: 'Connexion',
  error: 'Erreur de connexion',
  empty_displayname: 'Le nom d\'affichage ne peut pas être vide.',
}

export const context = {
  ignore: 'Ignorer',
  unignore: 'Ne plus ignorer',
  mute: 'Muet',
  unmute: 'Activer le son',
  release: 'Forcer la libération des contrôles',
  take: 'Forcer la prise des contrôles',
  give: 'Donner les contrôles',
  kick: 'Expulser',
  ban: 'Bannir l\'IP',
  confirm: {
    kick_title: 'Expulser {name} ?',
    kick_text: 'Êtes-vous sûr de vouloir expulser {name} ?',
    ban_title: 'Bannir {name} ?',
    ban_text: 'Voulez-vous bannir {name} ? Vous devrez redémarrer le serveur pour annuler.',
    mute_title: 'Mettre {name} en sourdine ?',
    mute_text: 'Êtes-vous sûr de vouloir mettre {name} en sourdine ?',
    unmute_title: 'Réactiver le son de {name} ?',
    unmute_text: 'Voulez-vous réactiver le son de {name} ?',
    button_yes: 'Oui',
    button_cancel: 'Annuler',
  },
}

export const controls = {
  release: 'Libérer les contrôles',
  request: 'Demander les contrôles',
  lock: 'Verrouiller les contrôles',
  unlock: 'Déverrouiller les contrôles',
  has: 'Vous avez le contrôle',
  hasnot: 'Vous n\'avez pas le contrôle',
  mic_on: 'Activer le microphone',
  mic_off: 'Désactiver le microphone',
  mic_error: 'Erreur de microphone',
}

export const locks = {
  control: {
    lock: 'Verrouiller les contrôles (pour les utilisateurs)',
    unlock: 'Déverrouiller les contrôles (pour les utilisateurs)',
    locked: 'Contrôles verrouillés (pour les utilisateurs)',
    unlocked: 'Contrôles déverrouillés (pour les utilisateurs)',
    notif_locked: 'a verrouillé les contrôles pour les utilisateurs',
    notif_unlocked: 'a déverrouillé les contrôles pour les utilisateurs',
  },
  login: {
    lock: 'Verrouiller la salle (pour les utilisateurs)',
    unlock: 'Déverrouiller la salle (pour les utilisateurs)',
    locked: 'Salle verrouillée (pour les utilisateurs)',
    unlocked: 'Salle déverrouillée (pour les utilisateurs)',
    notif_locked: 'a verrouillé la salle',
    notif_unlocked: 'a déverrouillé la salle',
  },
  file_transfer: {
    lock: 'Verrouiller le transfert de fichiers (pour les utilisateurs)',
    unlock: 'Déverrouiller le transfert de fichiers (pour les utilisateurs)',
    locked: 'Transfert de fichiers verrouillé (pour les utilisateurs)',
    unlocked: 'Transfert de fichiers déverrouillé (pour les utilisateurs)',
    notif_locked: 'a verrouillé le transfert de fichiers',
    notif_unlocked: 'a déverrouillé le transfert de fichiers',
  },
}

export const setting = {
  scroll: 'Sensibilité de défilement',
  scroll_invert: 'Inverser le défilement',
  trackpad_mode: 'Mode pavé tactile',
  trackpad_mode_description: 'Déplacer le curseur de façon relative comme un pavé tactile',
  trackpad_mode_mobile_hint: 'Deux doigts pour défiler, appuyer pour cliquer, maintenir pour clic droit',
  mobile: 'Mobile',
  desktop_only_inactive: 'Tactile uniquement',
  autoplay: 'Lecture automatique de la vidéo',
  ignore_emotes: 'Ignorer les émoticônes',
  chat_sound: 'Son du chat',
  keyboard_layout: 'Disposition du clavier',
  broadcast_title: 'Diffusion en direct',
}

export const connection = {
  logged_out: 'Vous avez été déconnecté.',
  reconnecting: 'Reconnexion...',
  connected: 'Connecté',
  disconnected: 'Déconnecté',
  kicked: 'Vous avez été expulsé de cette salle.',
  button_confirm: 'OK',
}

export const notifications = {
  connected: '{name} connecté',
  disconnected: '{name} déconnecté',
  controls_taken: '{name} a pris les contrôles',
  controls_taken_force: 'a pris les contrôles de force',
  controls_taken_steal: 'a pris les contrôles de {name}',
  controls_released: '{name} a libéré les contrôles',
  controls_released_force: 'a libéré les contrôles de force',
  controls_released_steal: 'a libéré les contrôles de {name}',
  controls_given: 'a donné les contrôles à {name}',
  controls_has: '{name} a les contrôles',
  controls_has_alt: 'Mais j\'ai fait savoir à la personne que vous le vouliez',
  controls_requesting: '{name} demande les contrôles',
  resolution: 'a changé la résolution en {width}x{height}@{rate}',
  banned: 'a banni {name}',
  kicked: 'a expulsé {name}',
  muted: 'a mis {name} en sourdine',
  unmuted: 'a réactivé le son de {name}',
}

export const files = {
  downloads: 'Téléchargements',
  uploads: 'Envois',
  upload_here: 'Cliquez ou faites glisser des fichiers ici pour les envoyer',
}
