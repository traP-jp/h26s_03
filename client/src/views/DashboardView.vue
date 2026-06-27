<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import type { components } from '../gen/api-types'
import { getPolls } from '../lib/api'

const router = useRouter()
//api/pollsの返り値
type Poll = components['schemas']['Poll']

const polls = ref<Poll[]>([])
// ページが表示されたときに投票一覧を取得する
async function loadPolls() {
  // const result = await fetch('/api/polls').then((r) => r.json())
  polls.value = await getPolls();
}

//投票の詳しい勝ち負け記入画面に行く
function goToPoll(id: number) {
  router.push(`/polls/${id}`)
}

//投票作成画面に行く
function goToCreatePoll() {
  router.push('/polls/create')
}
// ページが表示されたときに投票一覧を取得する
onMounted(() => {
  loadPolls()
})
// 投票一覧表示
</script>
<template>
  <div class="background">
    <div class="toppage">
      <h1>勝敗ギャンブル（仮）</h1>
      <h2>現在行われている投票</h2>
      <button v-for="poll in polls" :key="poll.id" class="poll-button" @click="goToPoll(poll.id)">
        <div class="poll-name">
          {{ poll.name }}
        </div>

      </button>

      <button class="more-button" @click="goToCreatePoll">
        もっと見る >
      </button>

      <button class="polladd-button" @click="goToCreatePoll">
        + 新しい投票を作成
      </button>
    </div>
  </div>
</template>

<style scoped>
.background {
  min-height: 100vh;
  display: flex;
  background-color: #0F172B;
  justify-content: center;
}

.toppage {
  min-height: 100vh;
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 70px;
  padding-bottom: 80px;
  background-color: #0f172b;
}

h1 {
  font-size: 40px;
  margin-bottom: 200px;
  font-weight: bold;
  color: #ffffff;
}

h2 {
  font-size: 20px;
  margin-bottom: 20px;
  color: #ffffff;
}

.poll-button {
  width: 420px;
  height: 50px;
  padding: 10px;
  margin-bottom: 10px;
  background: transparent;
  color: #ffffff;
  border: 2px solid #ffffff;
  border-radius: 5px;
  cursor: pointer;
}

.poll-button:hover {
  background: rgba(255, 255, 255, 0.103);
}

.more-button {
  width: 420px;
  height: 50px;
  padding: 10px;
  margin-top: 0px;
  background: transparent;
  color: #ffffff;
  border: 2px solid #ffffff58;
  border-radius: 5px;
  cursor: pointer;
}

.more-button:hover {
  background: rgba(255, 255, 255, 0.103);
}

.polladd-button {
  width: 420px;
  height: 50px;
  padding: 10px;
  margin-top: 60px;
  background: #193cb815;
  color: #ffffff;
  border: 2px solid #193CB8;

  border-radius: 5px;
  cursor: pointer;
}

.polladd-button:hover {
  background: rgba(248, 250, 255, 0.148);
}
</style>>