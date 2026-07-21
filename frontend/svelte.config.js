import adapter from "@sveltejs/adapter-node";
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),

	kit: {
		adapter: adapter({
<<<<<<< HEAD
			precompress: true,
=======
			precompress: false,
>>>>>>> bb3b73c (del precompress)
			trustProxy: true
		})
	}
};

export default config;
