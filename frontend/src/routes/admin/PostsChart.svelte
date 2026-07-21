<script lang="ts">
	import { t } from '$lib/translations';
	import { Line } from 'svelte-chartjs';
	import {
		Chart as ChartJS,
		CategoryScale,
		LinearScale,
		PointElement,
		LineElement,
		Title,
		Tooltip,
		Legend,
		Filler,
		type ChartOptions,
        type TooltipItem


	} from 'chart.js';

	ChartJS.register(
		CategoryScale,
		LinearScale,
		PointElement,
		LineElement,
		Title,
		Tooltip,
		Legend,
		Filler
	);

	interface BoardMetrics {
		alias: string;
		post_count: number;
		deleted_count: number;
		sage_count: number;
		thread_count: number;
	}

	interface DailyMetrics {
		date: string;
		boards: BoardMetrics[];
	}

	interface Props {
		dailyMetrics: DailyMetrics[];
	}

	let { dailyMetrics }: Props = $props();

	function generateColors(count: number): string[] {
		const colors: string[] = [];
		const goldenRatio = 0.618033988749895;
		let hue = Math.random() * 360;

		for (let i = 0; i < count; i++) {
			hue = (hue + goldenRatio * 360) % 360;
			const saturation = 55 + Math.random() * 20;
			const lightness = 45 + Math.random() * 10;
			colors.push(`hsl(${Math.round(hue)}, ${Math.round(saturation)}%, ${Math.round(lightness)}%)`);
		}
		return colors;
	}

	const boardAliases = $derived(
		[...new Set(dailyMetrics.flatMap((d) => d.boards.map((b) => b.alias)))].sort()
	);

	const colors = $derived(generateColors(boardAliases.length));

	const colorMap = $derived(
		Object.fromEntries(boardAliases.map((alias, i) => [alias, colors[i]]))
	);

	const labels = $derived(
		dailyMetrics.map((d) => {
			const date = new Date(d.date + 'T00:00:00');
			return date.toLocaleDateString(undefined, { weekday: 'short', month: 'short', day: 'numeric' });
		})
	);

	const maxCount = $derived(
		Math.max(...dailyMetrics.flatMap((d) => d.boards.map((b) => b.post_count)), 0)
	);

	const yStep = $derived(maxCount >= 10 ? 10 : 5);

	const yMax = $derived(Math.ceil((maxCount + yStep) / yStep) * yStep);

	const datasets = $derived(
		boardAliases.map((alias) => ({
			label: `/${alias}/`,
			data: dailyMetrics.map((day) => {
				const board = day.boards.find((b) => b.alias === alias);
				return board?.post_count ?? 0;
			}),
			borderColor: colorMap[alias],
			backgroundColor: colorMap[alias],
			fill: false,
			tension: 0.3,
			pointRadius: 4,
			pointHoverRadius: 6,
			borderWidth: 2
		}))
	);

	const chartData = $derived({
		labels,
		datasets
	});

	const options = $derived<ChartOptions<'line'>>({
		responsive: true,
		maintainAspectRatio: false,
		interaction: {
			mode: 'index' as const,
			intersect: false
		},
		plugins: {
			legend: {
				position: 'top' as const,
				labels: {
					usePointStyle: true,
					padding: 20,
					font: { size: 12 }
				}
			},
			tooltip: {
				backgroundColor: 'rgba(0, 0, 0, 0.8)',
				padding: 12,
				titleFont: { size: 13 },
				bodyFont: { size: 12 },
				callbacks: {
					label: (ctx: TooltipItem<"line">) => `${ctx.dataset.label}: ${ctx.parsed.y} posts`
				}
			}
		},
		scales: {
			x: {
				type: 'category',
				grid: { display: false },
				ticks: { font: { size: 11 } }
			},
			y: {
				type: 'linear',
				min: 0,
				max: yMax,
				ticks: {
					stepSize: yStep,
					font: { size: 11 },
					callback: (value: number | string) => Number(value)
				},
				grid: { color: 'rgba(0, 0, 0, 0.05)' },
				title: { display: true, text: $t('common.admin.dashboard.yAxis') }
			}
		}
	});
</script>

<div class="h-[400px] w-full">
	<Line data={chartData} options={options} />
</div>