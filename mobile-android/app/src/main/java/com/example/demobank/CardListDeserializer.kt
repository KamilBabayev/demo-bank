package com.example.demobank

import com.google.gson.JsonDeserializationContext
import com.google.gson.JsonDeserializer
import com.google.gson.JsonElement
import com.google.gson.reflect.TypeToken
import java.lang.reflect.Type

class CardListDeserializer : JsonDeserializer<List<Card>> {
    override fun deserialize(
        json: JsonElement?,
        typeOfT: Type?,
        context: JsonDeserializationContext?
    ): List<Card> {
        if (json?.isJsonArray == true) {
            // If the response is a JSON array, deserialize it as a list of cards
            val listType = object : TypeToken<List<Card>>() {}.type
            return context?.deserialize(json, listType) ?: emptyList()
        } else {
            // If the response is not an array (e.g., an object), return an empty list
            return emptyList()
        }
    }
}
