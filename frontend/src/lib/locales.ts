/** Supported UI locales and flag labels. */
export type AppLocale = 'en' | 'ru' | 'es';

export const APP_LOCALES: readonly {
	code: AppLocale;
	flag: string;
	label: string;
}[] = [
	{ code: 'en', flag: '🇺🇸', label: 'EN' },
	{ code: 'ru', flag: '🇷🇺', label: 'RU' },
	{ code: 'es', flag: '🇪🇸', label: 'ES' },
] as const;

export function isAppLocale(value: string): value is AppLocale {
	return APP_LOCALES.some((l) => l.code === value);
}

export function localeFlag(code: string): string {
	return APP_LOCALES.find((l) => l.code === code)?.flag ?? '🇺🇸';
}
