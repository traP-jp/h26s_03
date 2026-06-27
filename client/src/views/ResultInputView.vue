<script setup lang="ts">
import { onMounted, computed, ref } from "vue";
import { useRoute, useRouter, RouterLink } from "vue-router";

import { getPoll, type Poll, getMe, updatePoll } from "../lib/api";

const route = useRoute();
const router = useRouter();
const poll = ref<Poll | null>(null);
const me = ref();
const pollId = Number(route.params.id);
const selectedResult = ref<number | null>(null);
const loading = ref(true);
//結果が決まっているか判定
const finished = computed(() => {
  return poll.value?.result != null;
});
//編集権限があるか判定
const canEdit = computed(() => {
  return poll.value?.created_by === me.value?.username;
});
//選択を保存
const saveResult = async () => {
  if (selectedResult.value === null) {
    alert("結果を選択してください");
    return;
  }
  await updatePoll(pollId, selectedResult.value);
  router.push(`/polls/${pollId}`);
  console.log(selectedResult.value);
};

// 表示時に投票データを取得する
onMounted(async () => {
  const [pollData, meData] = await Promise.all([getPoll(pollId), getMe()]);
  poll.value = pollData;
  me.value = meData;
  loading.value = false;
  console.log(poll.value);
});

// 結果を選択する関数
const selectResult = (result: number) => {
  selectedResult.value = result;
};

//
</script>
<template>
  <div>
    <div class="background">
      <div class="result-input-page">
        <div v-if="loading" class="loading-component">
          <div class="neon-ring"></div>
          <p class="loading-text">読み込み中…</p>
        </div>
        <div v-else>
          <h1>{{ poll?.name ?? "読み込み中…" }}</h1>
          <p class="title-text">の投票を編集</p>
          <div v-if="finished">
            <div class="message-box">
              <h3>！編集できません！</h3>
              <p>この投票はすでに結果が確定しています</p>
            </div>
            <RouterLink class="return-button" :to="`/polls/${pollId}`"> 結果に戻る＞ </RouterLink>
          </div>
          <div v-else-if="!canEdit">
            <div class="message-box">
              <h3>！編集できません！</h3>
              <p>あなたには編集権限がありません</p>
            </div>
            <RouterLink class="return-button" :to="`/polls/${pollId}`"> 投票に戻る＞ </RouterLink>
          </div>
          <div v-else>
            <h2>勝った方を選択</h2>

            <div class="button-group">
              <button
                class="choice-button"
                @click="selectResult(1)"
                :class="{ selected: selectedResult === 1 }"
              >
                {{ poll?.choice1 }}
              </button>
              <button
                class="choice-button"
                @click="selectResult(2)"
                :class="{ selected: selectedResult === 2 }"
              >
                {{ poll?.choice2 }}
              </button>
            </div>

            <button class="save-button" @click="saveResult">決定</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.background {
  min-height: 100vh;
  display: flex;
  justify-content: center;
  background-color: #0f172b;
  justify-content: center;
}

.result-input-page {
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
  margin-bottom: -20px;
  color: #ffffff;
  text-align: center;
}

.title-text {
  font-size: 30px;
  transform: translateX(250px);
  color: #ffffff;
}
h2 {
  font-size: 20px;
  margin-bottom: 20px;
  transform: translateX(-5px);
  color: #ffffff;
}

.button-group {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-top: -20 px;
}

.choice-button {
  width: 420px;
  height: 50px;
  padding: 10px;
  margin-bottom: 10px;
  background: transparent;
  color: #ffffff;
  text-decoration: none;
  border: 2px solid #ffffff;
  border-radius: 5px;
  font-weight: bond;
  cursor: pointer;
}

.choice-button.selected {
  background-color: #155efc23;
  color: #155dfc;
  border: 2px solid #155dfc;
}
.choice-button:hover {
  background: rgba(255, 255, 255, 0.103);
}

.more-button:hover {
  background: rgba(255, 255, 255, 0.103);
}

.save-button {
  width: 420px;
  height: 50px;
  margin-top: 170px;
  background: #193cb815;
  color: #ffffff;
  text-decoration: none;
  text-align: center;
  border: 2px solid #193cb8;
  border-radius: 5px;
  cursor: pointer;

  transform: 0.2s;
}

.save-button:hover {
  background: rgba(132, 126, 255, 0.067);
}

.message-box {
  width: 480px;
  padding: 28px;
  margin-top: 50px;

  background: rgb(255, 247, 0);
  border: 2px solid #ef4444;
  border-radius: 10px;

  color: rgb(255, 255, 255);
  text-align: center;
}

.message-box h3 {
  font-size: 40px;
  margin-bottom: 14px;
  color: #f80000;
}

.message-box p {
  font-size: 19px;
  line-height: 1.6;
  color: #000;
}

.return-button {
  text-decoration: none;
  display: block;
  width: fit-content;
  color: #ffffff;
  margin: 20px auto 0;
}

.loading-component {
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.loading-component {
  width: 100%;
  height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  background: #0a0e27;
}

.neon-ring {
  width: 100px;
  height: 100px;
  border: 4px solid transparent;
  border-radius: 50%;
  border-top-color: #0ff;
  border-right-color: #0ff;
  position: relative;
  animation: spin 1.5s linear infinite;
  filter: drop-shadow(0 0 10px #0ff) drop-shadow(0 0 20px #0ff);
}

.neon-ring::before {
  content: "";
  position: absolute;
  inset: -8px;
  border-radius: 50%;
  border: 2px solid transparent;
  border-bottom-color: #f0f;
  border-left-color: #f0f;
  animation: spin 1s linear infinite reverse;
  filter: drop-shadow(0 0 10px #f0f) drop-shadow(0 0 20px #f0f);
}

@keyframes spin {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}

.loading-text {
  color: #cbd5e1;
  font-size: 18px;
  margin-bottom: 60px;
  letter-spacing: 1px;
}
</style>
