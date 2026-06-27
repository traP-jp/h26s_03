<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";

import { getPoll, type Poll, updatePoll } from "../lib/api";

const route = useRoute();
const router = useRouter();
const poll = ref<Poll | null>(null);
const pollId = Number(route.params.id);
const selectedResult = ref<number | null>(null);
const saveResult = async () => {
  if (selectedResult.value === null) {
    alert("勝った方を選択してください");
    return;
  }
  await updatePoll(pollId, selectedResult.value);
  router.push(`/polls/${pollId}`);
  console.log(selectedResult.value);
};

// 投票データを取得する
onMounted(async () => {
  poll.value = await getPoll(pollId);
});

// 結果を選択する関数
const selectResult = (result: number) => {
  selectedResult.value = result;
};
</script>
<template>
  <div>
    <div class="background">
      <div class="result-input-page">
        <h1>投票を編集</h1>
        <h2>勝った方を選択</h2>
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

        <button class="save-button" @click="saveResult">決定</button>
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
  transform: translateX(-130px);
  color: #ffffff;
}

.choice-button {
  width: 420px;
  height: 50px;
  padding: 10px;
  margin-bottom: 10px;
  background: transparent;
  color: #ffffff;
  text-decoration: none;
  text-align: center;
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
  padding: 10px;
  margin-top: 100px;
  background: #193cb815;
  color: #ffffff;
  text-decoration: none;
  text-align: center;
  border: 2px solid #193cb8;
  border-radius: 5px;
  cursor: pointer;
}

.save-button:hover {
  background: rgba(132, 126, 255, 0.067);
}
</style>
