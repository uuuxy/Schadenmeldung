/** Ein kurzes Antippen-Gefühl bei Erfolg (Android); Browser ohne Vibration überspringen es still. */
export const rueckmeldung = () => navigator.vibrate?.(12);
