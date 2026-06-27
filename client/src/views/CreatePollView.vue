<template>
  <div class="page-container">
    <div class="header">
      <router-link to="/" class="back-link">＜ 戻る</router-link>
      <h1>新しい投票を作成</h1>
    </div>
    <div class="form-container">
      <div class="form-group">
        <label for="poll-name" class="form-label">投票名</label>
        <input
          id="poll-name"
          v-model="form.name"
          type="text"
          class="form-input"
          placeholder="例) W杯 日本 vs チュニジア"
          required
        />
      </div>
      <div class="form-group">
        <label for="choice1" class="form-label">選択肢1</label>
        <input
          id="choice1"
          v-model="form.choice1"
          type="text"
          class="form-input"
          placeholder="例) 日本"
          required
        />
      </div>
      <div class="form-group">
        <label for="choice2" class="form-label">選択肢2</label>
        <input
          id="choice2"
          v-model="form.choice2"
          type="text"
          class="form-input"
          placeholder="例) チュニジア"
          required
        />
      </div>
      <div class="form-group">
        <label for="due" class="form-label">期限</label>
        <input id="due" v-model="dueModel" type="datetime-local" class="form-input" />
      </div>
      <div>
        <button type="button" class="submit-button" @click="submitForm">作成する</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive } from "vue";
import { useRouter } from "vue-router";

import { createPoll } from "../lib/api";
const router = useRouter();

type PollForm = {
  name: string;
  choice1: string;
  choice2: string;
  due: string | null;
};

const form = reactive<PollForm>({
  name: "",
  choice1: "",
  choice2: "",
  due: null,
});

const dueModel = computed({
  get: () => form.due ?? "",
  set: (value: string) => {
    form.due = value === "" ? null : value;
  },
});

console.log(form);

const submitForm = async () => {
  if (!form.name || !form.choice1 || !form.choice2) {
    alert("投票名と選択肢1, 2は必須です。");
    return;
  }
  if (form.choice1 === form.choice2) {
    alert("選択肢1と選択肢2は異なる値である必要があります。");
    return;
  }

  try {
    const res = await createPoll(form);
    if (res) {
      console.log(res);
      router.push(`/polls/${res.id}`);
      form.name = "";
      form.choice1 = "";
      form.choice2 = "";
      form.due = null;
    } else {
      throw new Error("投票の作成に失敗しました。");
    }
  } catch (error) {
    console.error(error);
    alert("エラーが発生しました。");
  }
};
</script>

<style scoped>
.header {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  margin-bottom: 20px;
}
.back-link {
  align-self: flex-start;
  margin: 10px;
  text-decoration: none;
  color: #ffffff;
}
.page-container {
  text-align: center;
  background-color: #0f172b;
  color: #ffffff;
  min-height: 100vh;
}
.form-container {
  max-width: 400px;
  margin: 0 auto;
  padding: 20px;
  border-radius: 8px;
}
.form-group {
  margin-bottom: 20px;
  display: flex;
  flex-direction: column;
  text-align: left;
}
.form-label {
  margin-bottom: 5px;
  font-weight: bold;
}
.form-input {
  padding: 10px;
  border: 2px solid #90a1b9;
  border-radius: 4px;
  width: 100%;
  background-color: #1d293d;
  color: #ffffff;
}
.form-input::placeholder {
  color: #45556c;
}

.submit-button {
  background-color: #162456;
  color: white;
  padding: 10px 20px;
  border: 1px solid #193cb8;
  border-radius: 4px;
  cursor: pointer;
}
.submit-button:hover {
  background-color: #193cb8;
}
</style>
