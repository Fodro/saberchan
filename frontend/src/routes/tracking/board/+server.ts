import type { VisitedBoard } from '$lib/types/metadata';
import type { RequestHandler } from '../$types';

export const POST: RequestHandler = async ({request, cookies}) => {
	const body: VisitedBoard = await request.json();
	const recentBoards: string[] = cookies.get('recent-boards')?.split(',') ?? [];
	
	const index = recentBoards.indexOf(body.alias);
	if (index !== -1) {
		recentBoards.splice(index, 1);
	}
	recentBoards.unshift(body.alias);

	while (recentBoards.length > 5){
		recentBoards.pop();
	}

	cookies.set('recent-boards', recentBoards.join(','), {path: '/'})

	return new Response('', {
		status: 200
	})
}; 