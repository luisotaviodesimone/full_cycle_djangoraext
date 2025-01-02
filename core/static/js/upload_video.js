const CHUNK_SIZE = 1 * 1024 * 1024; // 1MB per chunk
const MAX_SIMULTANEOUS_UPLOADS = 2;

document.getElementById('upload-form').addEventListener('submit', handleFormSubmit);

async function handleFormSubmit(event) {
  event.prevendDefault();
  changeSubmitStatus(true);
  const fileInput = event.target.querySelector('input[type="file"]');

  if (fileInput.files.legth === 0) {
    alert('Selecione um arquivo primeiro!');
    changeSubmitStatus(false);
    return;
  }

  try {
    const file = fileInput.files[0];
    await uploadFileInChunks(file);
  } finally {
    changeSubmitStatus(false);
  }

  function changeSubmitStatus(isDisabled) {
    const btnUploadVideo = document.getElementById('btnUploadVideo');
    btnUploadVideo.disabled = isDisabled;
    btnUploadVideo.value = isDisabled ? 'Enviando...' : 'Fazer upload';
  }

  async function uploadFileInChunks(file) {
    const totalChunks = Math.ceil(file.size / CHUNK_SIZE);

    const chunkPromises = generateChunkPromises(file, totalChunks);

    try {
      await runSimultaneousUploads(chunkPromises);
      await finishUpload(file.name, totalChunks);
    } catch (error) {
      displayError('Não foi possível fazer o upload do arquivo.');
      console.error(error);
    }
  }

  function generateChunkPromises(file, totalChunks) {
    let uploadedChunks = 0;
    const chunkPromises = [];
    for (let currentChunk = 0; currentChunk < totalChunks; currentChunk++) {
      const start = currentChunk * CHUNK_SIZE;
      const end = Math.min(start + CHUNK_SIZE, file.size);
      const chunk = file.slice(start, end);

      const formData = new FormData();
      formData.append('csrfmiddlewaretoken', document.querySelector('input[name="csrfmiddlewaretoken"]').value);
      formData.append('chunk', chunk);
      formData.append('chunkIndex', currentChunk);

      chunkPromises.push(() => uploadChunk(formData, () => {
        uploadedChunks++;
        updateProgress(uploadedChunks, totalChunks);
      }));
    }
    return chunkPromises;
  }

  async function uploadChunk(formData, onChunkUploaded) {
    try {
      const response = await fetch("{% url 'admin:core_video_upload' id=id %}", {
        method: 'POST',
        body: formData,
      });

      if (!response.ok) {
        const textError = await response.text();
        throw new Error(`Erro no upload do chunk: ${textError}`);
      }
      onChunkUploaded();  // Update progress after success
    } catch (error) {
      throw error;
    }
  }

  async function runSimultaneousUploads(uploadTasks) {
    const queue = uploadTasks.slice();  // Copy the list of uploads
    const activeUploads = [];

    while (queue.length > 0 || activeUploads.length > 0) {
      while (queue.length > 0 && activeUploads.length < MAX_SIMULTANEOUS_UPLOADS) {
        const task = queue.shift();
        const uploadPromise = task().finally(() => {
          activeUploads.splice(activeUploads.indexOf(uploadPromise), 1);
        });
        activeUploads.push(uploadPromise);
      }
      await Promise.race(activeUploads);
    }
  }

  async function finishUpload(fileName, totalChunks) {
    const formData = new FormData();
    formData.append('csrfmiddlewaretoken', document.querySelector('input[name="csrfmiddlewaretoken"]').value);
    formData.append('fileName', fileName);
    formData.append('totalChunks', totalChunks);

    const response = await fetch("{% url 'admin:core_video_upload_finish' id=id %}", {
      method: 'POST',
      body: formData,
    });

    if (!response.ok) {
      const textError = await response.text();
      throw new Error(`Erro ao finalizar o upload: ${textError}`);
    }

    window.location.href = "{% url 'admin:core_video_upload' id=id %}";
  }

  function updateProgress(uploadedChunks, totalChunks) {
    const progressElement = document.getElementById('progress');
    const percentage = Math.floor((uploadedChunks / totalChunks) * 100);
    progressElement.innerText = `${percentage}%`;
    progressElement.style.width = `${percentage}%`;
  }

  function displayError(message) {
    const errorNoteElement = document.getElementById('errornote');
    errorNoteElement.classList.remove('hidden');
    errorNoteElement.innerText = message;
  }
}

