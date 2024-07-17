import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import dotenv from 'dotenv';

dotenv.config();

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		host: process.env.VITE_HOST || '0.0.0.0',
		port: parseInt(process.env.VITE_PORT) || 3000
	}
});
