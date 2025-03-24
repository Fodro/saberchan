import i18n from 'sveltekit-i18n';

/** @type {import('sveltekit-i18n').Config} */
const config = ({
	loaders: [
		{
			locale: 'en',
			key: 'common',
			loader: async () => (
				await import('./en/common.json')
			).default,
		},
		{
			locale: 'ru',
			key: 'common',
			loader: async () => (
				await import('./ru/common.json')
			).default,
		},
	],
});

export const { t, locale, locales, loading, loadTranslations, addTranslations, translations, setLocale, setRoute } = new i18n(config);