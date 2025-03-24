import { addTranslations, setLocale, setRoute } from "$lib/translations";
import type { LayoutLoad } from "./$types";

export const load: LayoutLoad = async ({ data }) => {
	const { i18n, translations } = data;
	const { locale, route } = i18n;

	addTranslations(translations);

	await setRoute(route);
	await setLocale(locale);
	
	return {
		...data
	}
}