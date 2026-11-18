import { APIPath, HTTPMethod } from "~/libs/enums/enums.js";
import { configureString } from "~/libs/helpers/helpers.js";
import { baseApi } from "~/libs/modules/api/base-api.js";

import { NotesApiPath, NotesApiTag } from "./libs/enums/enums.js";
import { type NoteCreateRequestDto, type NoteDto } from "./libs/types/types.js";

const notesApi = baseApi
	.enhanceEndpoints({ addTagTypes: [NotesApiTag.NOTE] })
	.injectEndpoints({
		endpoints: (builder) => ({
			createNote: builder.mutation<NoteDto, NoteCreateRequestDto>({
				extraOptions: { shouldSuppressToast: true },
				invalidatesTags: [NotesApiTag.NOTE],
				query: (payload) => ({
					body: payload,
					method: HTTPMethod.POST,
					url: APIPath.NOTES,
				}),
			}),

			deleteNote: builder.mutation<null, number>({
				invalidatesTags: (_result, error) => {
					const hasError = Boolean(error);

					return hasError ? [] : [NotesApiTag.NOTE];
				},
				query: (id) => ({
					method: HTTPMethod.DELETE,
					url: configureString(APIPath.NOTES, NotesApiPath.$ID, {
						id: String(id),
					}),
				}),
			}),

			getNotes: builder.query<NoteDto[], undefined>({
				providesTags: [NotesApiTag.NOTE],
				query: () => ({
					url: APIPath.NOTES,
				}),
			}),
		}),
	});

const { useCreateNoteMutation, useDeleteNoteMutation, useGetNotesQuery } =
	notesApi;

export { useCreateNoteMutation, useDeleteNoteMutation, useGetNotesQuery };
