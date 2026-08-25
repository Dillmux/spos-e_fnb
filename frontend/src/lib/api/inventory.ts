import { request } from './client';
import type { Ingredient, ProductRecipe } from '$lib/types';

export const inventoryApi = {
  getIngredients: async () => {
    return request<Ingredient[]>('/inventory/ingredients');
  },

  saveIngredient: async (data: Partial<Ingredient>) => {
    const method = data.id ? 'PUT' : 'POST';
    const url = data.id ? `/inventory/ingredients/${data.id}` : '/inventory/ingredients';
    const payload = {
      name: data.name,
      unit: data.unit,
      current_stock: data.current_stock ?? data.stock ?? 0,
      min_stock_alert: data.min_stock_alert ?? data.min_stock ?? 0,
      cost_per_unit: data.cost_per_unit ?? data.cogs_per_unit ?? 0,
    };
    return request<Ingredient>(url, {
      method,
      body: JSON.stringify(payload),
    });
  },

  deleteIngredient: async (id: string) => {
    return request<{ id: string }>(`/inventory/ingredients/${id}`, {
      method: 'DELETE',
    });
  },

  adjustStock: async (
    ingredientId: string,
    payloadOrDelta: number | { change_type?: string; quantity_change?: number; delta?: number; notes?: string; reason?: string },
    reasonStr: string = ''
  ) => {
    let changeType = 'restock';
    let qtyChange = 0;
    let notes = '';

    if (typeof payloadOrDelta === 'number') {
      qtyChange = payloadOrDelta;
      changeType = qtyChange < 0 ? 'waste' : 'restock';
      notes = reasonStr || 'Stock adjustment';
    } else {
      qtyChange = payloadOrDelta.quantity_change ?? payloadOrDelta.delta ?? 0;
      changeType = payloadOrDelta.change_type || (qtyChange < 0 ? 'waste' : 'restock');
      notes = payloadOrDelta.notes || payloadOrDelta.reason || reasonStr || 'Stock adjustment';
    }

    return request<{ ingredient_id: string; new_stock: number; quantity_change: number; change_type: string }>(
      `/inventory/ingredients/${ingredientId}/adjust`,
      {
        method: 'POST',
        body: JSON.stringify({
          change_type: changeType,
          quantity_change: qtyChange,
          notes,
        }),
      }
    );
  },

  getRecipe: async (productId: string) => {
    return request<ProductRecipe[]>(`/inventory/recipes/${productId}`);
  },

  saveRecipe: async (productId: string, recipes: ProductRecipe[]) => {
    const formattedRecipes = recipes.map((r) => ({
      ingredient_id: r.ingredient_id,
      quantity_required: r.quantity_required ?? r.quantity ?? 0,
    }));
    return request<{ product_id: string; calculated_cogs: number; recipes_count: number }>(`/inventory/recipes/${productId}`, {
      method: 'POST',
      body: JSON.stringify({ recipes: formattedRecipes }),
    });
  },

  getLogs: async () => {
    return request<any[]>('/inventory/logs');
  },
};

