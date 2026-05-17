import { objectHasKey, parseJSON } from '@/utils/helper';
import { sendMessage } from '@/utils/message';
import { openDB } from 'idb';
import deepmerge from 'lodash.merge';
import { nanoid } from 'nanoid';
import browser from 'webextension-polyfill';

function initWebListener() {
  const listeners = {};

  function on(name, callback) {
    (listeners[name] = listeners[name] || []).push(callback);
  }

  window.addEventListener('__automa-ext__', ({ detail }) => {
    if (!detail || !objectHasKey(listeners, detail.type)) return;

    listeners[detail.type].forEach((listener) => {
      listener(detail.data);
    });
  });

  return { on };
}
function sendMessageBack(type, payload = {}) {
  const event = new CustomEvent(`__automa-ext__${type}`, {
    detail: payload,
  });

  window.dispatchEvent(event);
}

// BrowserFlow local change start: forward workflow execution result to page 转发工作流执行结果到页面
browser.runtime.onMessage.addListener((message) => {
  if (message?.type !== 'browserflow:workflow-result') return undefined;

  window.dispatchEvent(
    new CustomEvent('__browserflow_automa_workflow_result__', {
      detail: message.data || {},
    })
  );
  return undefined;
});
// BrowserFlow local change end


// BrowserFlow local change start: preserve imported workflow timestamps 保留导入工作流的原始时间戳
// normalizeWorkflowTimestamp keeps imported Automa timestamps 保留导入工作流的原始时间
function normalizeWorkflowTimestamp(value, fallbackValue = Date.now()) {
  const numberValue = Number(value);
  if (Number.isFinite(numberValue) && numberValue > 0) return numberValue;

  const dateValue = new Date(value).getTime();
  if (!Number.isNaN(dateValue)) return dateValue;

  return fallbackValue;
}
// BrowserFlow local change end

async function initWebServiceBridge() {
  try {
    document.body.setAttribute(
      'data-atm-ext-installed',
      browser.runtime.getManifest().version
    );

    const { workflows } = await browser.storage.local.get('workflows');
    const db = await openDB('automa', 1, {
      upgrade(event) {
        event.createObjectStore('store');
      },
    });

    await db.put('store', workflows, 'workflows');

    const webListener = initWebListener();
    webListener.on('open-dashboard', ({ path }) => {
      if (!path) return;

      sendMessage('open:dashboard', path, 'background');
    });
    webListener.on('open-workflow', ({ workflowId }) => {
      if (!workflowId) return;

      sendMessage('open:dashboard', `/workflows/${workflowId}`, 'background');
    });
    // BrowserFlow local change start: import server workflow with stable id and ack 带稳定 ID 导入服务端工作流并回执
    webListener.on('add-workflow', async ({ workflow, requestId }) => {
      try {
        const { workflows: storedWorkflows } = await browser.storage.local.get(
          'workflows'
        );
        const workflowsStorage = storedWorkflows || {};

        const workflowId = workflow.id || nanoid();
        const now = Date.now();
        const workflowData = {
          ...workflow,
          id: workflowId,
          dataColumns: workflow.dataColumns || [],
          createdAt: normalizeWorkflowTimestamp(workflow.createdAt, now),
          updatedAt: normalizeWorkflowTimestamp(workflow.updatedAt, now),
          table: workflow.table || workflow.dataColumns || [],
        };

        workflowData.drawflow =
          typeof workflowData.drawflow === 'string'
            ? parseJSON(workflowData.drawflow, workflowData.drawflow)
            : workflowData.drawflow;

        if (Array.isArray(workflowsStorage)) {
          const workflowIndex = workflowsStorage.findIndex(
            (item) => item.id === workflowId
          );

          if (workflowIndex === -1) {
            workflowsStorage.push(workflowData);
          } else {
            workflowsStorage[workflowIndex] = workflowData;
          }
        } else {
          workflowsStorage[workflowId] = workflowData;
        }

        await browser.storage.local.set({ workflows: workflowsStorage });
        sendMessage(
          'workflow:added',
          { workflowId, workflowData },
          'background'
        );
        sendMessageBack('add-workflow', {
          ok: true,
          requestId,
          workflow: workflowData,
          workflow_id: workflowId,
          name: workflowData.name,
        });
      } catch (error) {
        console.error(error);
        sendMessageBack('add-workflow', {
          ok: false,
          requestId,
          error: error.message,
        });
      }
    });
    // BrowserFlow local change end
    webListener.on('add-team-workflow', async ({ workflow }) => {
      let { teamWorkflows } = await browser.storage.local.get('teamWorkflows');

      let workflowData = {
        ...workflow,
        createdAt: Date.now(),
        table: workflow.table ?? [],
      };
      workflowData.drawflow =
        typeof workflowData.drawflow === 'string'
          ? parseJSON(workflowData.drawflow, workflowData.drawflow)
          : workflowData.drawflow;

      if (!teamWorkflows) teamWorkflows = {};
      if (!teamWorkflows[workflowData.teamId])
        teamWorkflows[workflowData.teamId] = {};

      const workflowToMerge =
        teamWorkflows[workflowData.teamId][workflow.id] || null;
      if (workflowToMerge) {
        workflowData = deepmerge(workflowToMerge, workflowData);
      }

      teamWorkflows[workflowData.teamId][workflow.id] = workflowData;
      await browser.storage.local.set({ teamWorkflows });

      const triggerBlock = workflowData.drawflow.nodes?.find(
        (node) => node.label === 'trigger'
      );
      if (triggerBlock) {
        await sendMessage(
          'workflow:register',
          { triggerBlock, workflowId: workflowData.id },
          'background'
        );
      }

      sendMessage(
        'workflow:added',
        {
          workflowId: workflowData.id,
          teamId: workflowData.teamId,
          source: 'team',
        },
        'background'
      );
    });
    webListener.on('check-team-workflow', async ({ teamId, workflowId }) => {
      const { teamWorkflows } = await browser.storage.local.get(
        'teamWorkflows'
      );
      const workflowExist = Boolean(teamWorkflows?.[teamId]?.[workflowId]);

      window.dispatchEvent(
        new CustomEvent('__automa-team-workflow__', {
          detail: { exists: workflowExist },
        })
      );
    });
    webListener.on('add-package', async (data) => {
      try {
        const { savedBlocks } = await browser.storage.local.get('savedBlocks');
        const packages = savedBlocks || [];

        packages.push({ ...data.package, createdAt: Date.now() });

        await browser.storage.local.set({ savedBlocks: packages });

        sendMessage('dashboard:refresh-packages', '', 'background');
      } catch (error) {
        console.error(error);
      }
    });
    webListener.on('update-package', async (data) => {
      const { savedBlocks } = await browser.storage.local.get('savedBlocks');
      const packages = savedBlocks || [];

      const index = packages.findIndex((pkg) => pkg.id === data.id);
      if (index === -1) return;

      Object.assign(packages[index], data.package);

      await browser.storage.local.set({ savedBlocks: packages });

      sendMessage('dashboard:refresh-packages', '', 'background');
    });
    webListener.on('send-message', async ({ type, data }) => {
      if (type === 'package-installed') {
        const { savedBlocks } = await browser.storage.local.get('savedBlocks');
        const packages = savedBlocks || [];
        const isInstalled = packages.some((pkg) => pkg.id === data);

        sendMessageBack(type, isInstalled);
      } else if (type === 'get-workflows') {
        const storage = await browser.storage.local.get('workflows');
        sendMessageBack(type, storage.workflows);
      }
    });
  } catch (error) {
    console.error(error);
  }
}

// BrowserFlow local change start: initialize bridge after late extension injection 支持扩展延迟注入后初始化桥接
if (document.readyState === 'loading') {
  window.addEventListener('DOMContentLoaded', initWebServiceBridge, {
    once: true,
  });
} else {
  initWebServiceBridge();
}
// BrowserFlow local change end

window.addEventListener('user-logout', () => {
  browser.storage.local.remove(['session', 'sessionToken']);
});

window.addEventListener('app-mounted', async () => {
  try {
    const STORAGE_KEY = 'supabase.auth.token';
    const webStorageAuthData = parseJSON(
      localStorage.getItem(STORAGE_KEY),
      null
    );
    const extensionStorage = await browser.storage.local.get([
      'session',
      'sessionToken',
    ]);

    const setUserSession = async () => {
      const saveToStorage = { session: webStorageAuthData };

      const isGoogleProvider =
        webStorageAuthData?.user?.user_metadata?.iss.includes('google.com');
      const { session: currSession, sessionToken: currSessionToken } =
        await browser.storage.local.get(['session', 'sessionToken']);
      if (
        isGoogleProvider &&
        ((webStorageAuthData &&
          webStorageAuthData.user.id === currSession?.user.id) ||
          !currSessionToken)
      ) {
        saveToStorage.sessionToken = {
          access: webStorageAuthData.provider_token,
          refresh: webStorageAuthData.provider_refresh_token,
        };
      }
      if (!isGoogleProvider) {
        browser.storage.local.remove('sessionToken');
      }

      await browser.storage.local.set(saveToStorage);
    };

    if (webStorageAuthData && !extensionStorage.session) {
      await setUserSession();
    } else if (webStorageAuthData && extensionStorage.session) {
      if (webStorageAuthData.user.id !== extensionStorage.session.id) {
        await setUserSession();
      } else {
        const currentSession = { ...extensionStorage.session };
        if (extensionStorage.sessionToken) {
          currentSession.provider_token = extensionStorage.sessionToken.access;
          currentSession.provider_refresh_token =
            extensionStorage.sessionToken.refresh;
        }

        localStorage.setItem(STORAGE_KEY, JSON.stringify(currentSession));
      }
    }
  } catch (error) {
    console.error(error);
  }
});
