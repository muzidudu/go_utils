<script lang="ts">
	import { page } from "$app/state";
	import { Button } from "$lib/components/ui/button";
	let { children } = $props();
</script>

<div class="flex min-h-screen bg-slate-50">
	<!-- 左侧边栏 -->
	<aside class="w-56 border-r border-slate-200 bg-white">
		<div class="flex h-14 items-center border-b border-slate-200 px-6">
			<a href="/" class="text-lg font-semibold text-slate-900">后台管理</a
			>
		</div>
		<nav class="flex flex-col gap-1 p-4">
			<a
				href="/sites"
				class="rounded-md px-3 py-2 text-sm font-medium transition-colors {page.url.pathname.startsWith(
					'/sites',
				)
					? 'bg-slate-100 text-slate-900'
					: 'text-slate-600 hover:bg-slate-50 hover:text-slate-900'}"
			>
				站点管理
			</a>
			<a
				href="/categories"
				class="rounded-md px-3 py-2 text-sm font-medium transition-colors {page.url.pathname.startsWith(
					'/categories',
				)
					? 'bg-slate-100 text-slate-900'
					: 'text-slate-600 hover:bg-slate-50 hover:text-slate-900'}"
			>
				分类管理
			</a>
		</nav>
	</aside>

	<!-- 右侧内容区 -->
	<main class="flex-1 overflow-auto">
		<div class="container mx-auto max-w-6xl p-6">
			{@render children()}
		</div>
	</main>
</div>
