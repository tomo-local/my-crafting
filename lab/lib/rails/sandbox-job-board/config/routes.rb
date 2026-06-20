Rails.application.routes.draw do
  get "users/new"
  resources :companies, only: [:index, :show, :create]
  resources :users, only: [:index, :show, :create]
  resources :jobs, only: [:index, :show, :create] do
    resources :entries, only: [:create]
  end

  get "up" => "rails/health#show", as: :rails_health_check
  # Defines the root path route ("/")
  # root "posts#index"
end
