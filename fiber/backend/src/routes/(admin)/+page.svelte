<script lang="ts">
</script>

<svelte:head>
	<title>后台管理</title>
</svelte:head>

<div class="space-y-6">
	<h1 class="text-2xl font-bold text-slate-900">欢迎使用后台管理</h1>
	<p class="text-slate-600">请从左侧菜单选择功能。</p>
	<div class="flex gap-4">
		<a
			href="/sites"
			class="rounded-lg border border-slate-200 bg-white px-4 py-3 text-sm font-medium text-slate-700 shadow-sm transition-colors hover:bg-slate-50"
		>
			站点管理
		</a>
		<a
			href="/categories"
			class="rounded-lg border border-slate-200 bg-white px-4 py-3 text-sm font-medium text-slate-700 shadow-sm transition-colors hover:bg-slate-50"
		>
			分类管理
		</a>
	</div>
</div>
