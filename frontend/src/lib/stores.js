import { writable } from 'svelte/store';

export const channels = writable([]);
export const selectedChannel = writable(null);
export const items = writable([]);
export const selectedItem = writable(null);
