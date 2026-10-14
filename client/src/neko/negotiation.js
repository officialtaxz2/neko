export function shouldCreateClientOffer({
  currentPeer,
  socketOpen,
  hasRemoteDescription,
  signalingState,
  makingOffer,
}) {
  return (
    currentPeer &&
    socketOpen &&
    hasRemoteDescription &&
    signalingState === 'stable' &&
    !makingOffer
  )
}
