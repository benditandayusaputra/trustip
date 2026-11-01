import { createContext } from 'svelte';
import type { AccountSession } from './session.svelte';

export interface SessionSlot {
	session?: AccountSession;
}

export const [getSessionSlot, setSessionSlot, hasSessionSlot] = createContext<SessionSlot>();
